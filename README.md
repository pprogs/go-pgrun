# GO-PGRUN

Simple CLI tool to run sql commands from files against postgresql

Command line flags:
```
  -C string
        Path to config file (default "config.json")
  -D string
        Path to data file
  -V value
        Set value. Must be in format -V name,value
```

-V can be used multiple times like
```
go-pgrun -D commands.sql -V dbname,dev -V needver,0.1.4
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
\val dbname devDb
\val user devUser
\val ver 0.1.1

\os windows
\val collate Russian_Russia.1251
\os linux
\val collate ru-Ru.utf8
\os

DROP DATABASE IF EXISTS ##dbname##;
\go

CREATE DATABASE ##dbname##
    WITH 
    OWNER = ##user##
    ENCODING = 'UTF8'
    LC_COLLATE = '##collate##' 
    LC_CTYPE = '##collate##'
    TABLESPACE = pg_default
    CONNECTION LIMIT = -1
    TEMPLATE template0;
\go

--reconnect to db
\db ##dbname##

--check for db version
--breaks execution if not equal
--version is stored as comment on database 
--\needVer ##ver##

\os windows
COMMENT ON DATABASE ##dbname## IS '##ver##';
\os linux
COMMENT ON DATABASE ##dbname## IS 'im on linux';
\os
```

Versions must follow [SemVer](http://semver.org/).
