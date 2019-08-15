# GO-PGRUN

Simple CLI tool to run sql commands from files against postgresql

Command line flags:
```
-C string
    Path to config file (default "config.json")
-D string
    Path to data file
```

Simple config.json:
```
{
    "user" : "postgres",
    "password" : "password",
    "database" : "postgres"
}
```

Simple data file:
```sql
\val dbname remdesk
\val nedver 0.1.3
\val curver 0.1.4

--reconnect to another db
\db ##dbname##

--check for db version
--breaks execution if not equal
--version is stored as comment on database 
\needVer ##nedver##

\os windows
COMMENT ON DATABASE ##dbname## IS '##curver##';
\os linux
COMMENT ON DATABASE ##dbname## IS '0.2.3';
\os

--run batch
\go
```

Versions must follow [SemVer](http://semver.org/).
