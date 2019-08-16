\val dbname remdesk
\val nedver 0.1.4
\val curver 0.1.4

--reconnect to another db
\db ##dbname##

--check for db version
--breaks execution if not equal
--version is stored as comment on database 
\needVer ##needver##

\os windows
COMMENT ON DATABASE ##dbname## IS '##curver##';
\os linux
COMMENT ON DATABASE ##dbname## IS '0.2.3';
\os

--run batch
\go


