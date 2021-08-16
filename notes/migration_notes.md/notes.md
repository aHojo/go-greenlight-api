# Use for migrations
SQL migrations files for our database.

```sql
postgres=# CREATE DATABASE greenlight;
CREATE DATABASE
postgres=# \c greenlight
You are now connected to database "greenlight" as user "postgres".
greenlight=# CREATE ROLE greenlight WITH LOGIN PASSWORD 'password';
CREATE ROLE
greenlight=# CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION
greenlight=#
```

generate a config file
https://pgtune.leopard.in.ua/#/

important psql settings
https://www.enterprisedb.com/postgres-tutorials/how-tune-postgresql-memory

find the settings file
```sql
postgres@6ace5fb2d64b:/$ psql -c 'SHOW config_file;'
               config_file
------------------------------------------
 /var/lib/postgresql/data/postgresql.conf
(1 row)
```

# Go driver for postgres
`go get github.com/lib/pq@v1.10.0`

DSN
`postgres://username:password@localhost/greenlight`

SET AN ENVIRONMENT VARIABLE
`export GREENLIGHT_DB_DSN='postgres://username:password@localhost/greenlight?sslmode=disable'`