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

# Migration tool
We will use the migrate command line tool. 
[migrate](https://github.com/golang-migrate/migrate)

```
 curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz

  mv migrate.linux-amd64 $GOPATH/bin/migrate
 ```

 ## Create the migration files
 `migrate create -seq -ext=.sql -dir=./migrations create_movies_table`

`-seq` flag indicates that we want to use sequential numbering like 0001, 0002, for the migration files (instead of a Unix timestamp, which is the default).  
`-ext` flag indicates that we want to give the migration files the extension .sql.  
`-dir` flag indicates that we want to store the migration files in the ./migrations directory (which will be created automatically if it doesn’t already exist).  
The name  `create_movies_table` is a descriptive label that we give the migration files to
signify their contents.

**migrations folder now** 
```
➜  go-greenlight-api git:(sql_migrations) ✗ ls -l migrations 
total 0
-rw-r--r-- 1 ahojo ahojo 0 Aug 16 15:39 000001_create_movies_table.down.sql
-rw-r--r-- 1 ahojo ahojo 0 Aug 16 15:39 000001_create_movies_table.up.sql
```

`migrate create -seq -ext=.sql -dir=./migrations add_movies_check_constraints`

## Apply the migrations 
`migrate -path=./migrations -database=$GREENLIGHT_DB_DSN up`