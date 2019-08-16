package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-pg/pg"
	"github.com/hashicorp/go-version"

	sutils "github.com/pprogs/simpleutils"
)

type configFile struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

var (
	db    *pg.DB
	dbOpt *pg.Options
	err   error
	conf  *configFile
	data  []byte
	lg    *log.Logger
	valRx *regexp.Regexp
	sync  chan bool
	vals  valFlag
)

func main() {
	os.Exit(mainFunc())
}

func mainFunc() int {

	lg = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	confFlag := flag.String("C", "config.json", "Path to config file")
	dataFlag := flag.String("D", "", "Path to data file")

	flag.Var(&vals, "V", "Set value. Must be in format -V name,value")

	flag.Parse()

	if *confFlag == "" || *dataFlag == "" {
		flag.PrintDefaults()
		return 1
	}

	confFile := *confFlag
	dataFile := *dataFlag

	conf = &configFile{}
	if _, err = sutils.ReadJSON(confFile, conf); err != nil {
		lg.Printf("Could not read CONFIG file from %s!\n", confFile)
		return 1
	}

	if data, err = sutils.ReadFileData(dataFile); err != nil {
		lg.Printf("Could not read DATA file from %s!\n", dataFile)
		return 1
	}

	dbOpt = &pg.Options{
		User:     conf.User,
		Password: conf.Password,
		Database: conf.Database,
	}

	db = pg.Connect(dbOpt)

	if db == nil {
		lg.Printf("Could not open DB connection!\n")
		return 1
	}

	defer func() {
		if db != nil {
			lg.Printf("Closing db connection")
			db.Close()
		}
	}()

	sync = make(chan bool)
	defer close(sync)

	commands := string(data)
	var tx *pg.Tx

	for batch := range generateBatches(commands) {

		lg.Printf("running batch: %s\n", batch)

		if tx, err = db.Begin(); err != nil {
			break
		}

		if _, err = db.Exec(batch); err != nil {
			tx.Rollback()
			break
		}

		tx.Commit()
		sync <- true
	}

	if err != nil {
		lg.Printf("Error:%+v\n", err)
		return 1
	}

	lg.Printf("Done without errors!\n")
	return 0
}

func checkVer(verStr string) int {

	var ver *version.Version
	var pgVer *version.Version

	ver, err = version.NewVersion(verStr)
	if err != nil {
		lg.Printf("Cannot parse version (%s)\n", verStr)
		return 2
	}

	q := `SELECT pg_catalog.shobj_description(d.oid, 'pg_database') AS "Description"
			FROM   pg_catalog.pg_database d
			WHERE  datname = current_database();`

	pgVerStr := ""
	if _, err = db.Query(pg.Scan(&pgVerStr), q); err != nil {
		lg.Printf("Cannot get version from PG\n")
		return 2
	}

	pgVer, err = version.NewVersion(pgVerStr)
	if err != nil {
		lg.Printf("Cannot parse PG version (%s)\n", pgVerStr)
		return 2
	}

	if !pgVer.Equal(ver) {
		lg.Printf("PG ver %s != file ver %s\n", pgVerStr, verStr)
		return 1
	}

	lg.Printf("Version ok\n")
	return 0
}

func generateBatches(input string) <-chan string {

	c := make(chan string)

	valRx = regexp.MustCompile(`^\\(\w+)\s*([\w#\d.]+)?\s*([\w#\d.]+)?$`)
	repl := vals.Replacer()

	runBatch := func(b *strings.Builder) bool {
		if b.Len() > 0 {
			//send batch
			c <- b.String()
			b.Reset()
			//wait for batch to complete
			if _, ok := <-sync; !ok {
				return false
			}
		}
		return true
	}

	go func() {

		defer close(c)

		// split input
		strAr := strings.Split(strings.Replace(input, "\r\n", "\n", -1), "\n")
		b := strings.Builder{}
		skip := false

		for idx := range strAr {

			s := strings.TrimSpace(strAr[idx])

			if len(s) == 0 {
				continue
			}

			//replace vals
			if len(vals) > 0 {
				s = repl.Replace(s)
			}

			//skip comments
			if strings.HasPrefix(s, "--") {
				continue
			}

			//check for commands
			if strings.HasPrefix(s, "\\") {

				v := valRx.FindAllStringSubmatch(s, -1)

				if len(v) == 0 {
					continue
				}

				command := strings.ToLower(v[0][1])
				val1 := ""
				val2 := ""
				if len(v[0]) > 1 {
					val1 = v[0][2]
				}
				if len(v[0]) > 2 {
					val2 = v[0][3]
				}

				lg.Printf("Running command %s (%s) (%s)\n", command, val1, val2)

				//os check
				if command == "os" && val1 != "" {
					if strings.ToLower(val1) != runtime.GOOS {
						skip = true
					} else {
						skip = false
					}
					continue
				}

				//stop os check
				if command == "os" && val1 == "" {
					skip = false
					continue
				}

				//os check also skips commands
				if skip {
					continue
				}

				//run batch
				if command == "go" {
					if !runBatch(&b) {
						return
					}
					continue
				}

				//set vals
				if command == "val" && val1 != "" && val2 != "" {
					vals.Add(val1, val2)
					repl = vals.Replacer()
					continue
				}

				//check version
				if command == "needver" && val1 != "" {
					ret := checkVer(val1)
					if ret == 2 {
						err = errors.New("error parsing version")
						return
					}
					if ret == 1 {
						err = errors.New("wrong version")
						return
					}
					continue
				}

				//db
				if command == "db" && val1 != "" {
					lg.Printf("Reconnect to database (%s)\n", val1)
					db.Close()
					dbOpt.Database = val1
					db = pg.Connect(dbOpt)
					continue
				}

				lg.Printf("Unknown command %s\n", command)

				continue
			}

			//add to batch
			if !skip {
				if b.Len() > 0 {
					b.WriteRune('\n')
				}
				b.WriteString(s)
			}

		}

		runBatch(&b)
	}()

	return c
}
