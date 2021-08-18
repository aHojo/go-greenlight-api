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

## Add indexes to our database
```sql
➜  go-greenlight-api git:(FilterSortPagination) ✗ export GREENLIGHT_DB_DSN='postgres://username:password@localhost/greenlight?sslmode=disable'

➜  go-greenlight-api git:(FilterSortPagination) migrate create -seq -ext .sql -dir ./migrations add_movies_indexes
/home/ahojo/development/go/src/go-greenlight-api/migrations/000003_add_movies_indexes.up.sql
/home/ahojo/development/go/src/go-greenlight-api/migrations/000003_add_movies_indexes.down.sql

➜  go-greenlight-api git:(FilterSortPagination) ✗ migrate -path ./migrations -database $GREENLIGHT_DB_DSN up
3/u add_movies_indexes (16.782845ms)
```

## Creating the Users Tabel
```sql
 migrate create -se
q -ext=.sql -dir=./migrations create_users_table
/home/ahojo/development/go/src/go-greenlight-api/migrations/000004_create_users_table.up.sql
/home/ahojo/development/go/src/go-greenlight-api/migrations/000004_create_users_table.down.sql
```
04 up
```sql
CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  create_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  name text NOT NULL,
  email citext UNIQUE NOT NULL,
  password_hash bytea NOT NULL,
  activated bool NOT NULL,
  version integer NOT NULL DEFAULT 1
);
```

04 down 
```sql
DROP TABLE IF EXISTS users;
```

1.  The email column has the type citext (case-insensitive text). This type stores text data exactly as it is inputted — without changing the case in any way — but comparisons against the data are always case-insensitive... including lookups on associated indexes.

2.  We’ve also got a `UNIQUE` constraint on the  `email` column. Combined with the `citext` type, this means that no two rows in the database can have the same `email` value — even if they have different cases. This essentially enforces a database-level business rule that no two users should exist with the same `email` address.

3.  The `password_hash` column has the type `bytea` (binary string). In this column we’ll store
a one-way hash of the user’s password generated using `bcrypt` — not the plaintext password

4.  The activated column stores a `boolean` value to denote whether a user account is ‘active’ or not. We will set this to false by default when creating a new user, and require the user to confirm their email address before we set it to true.

5.  We’ve also included a version number column, which we will increment each time a user record is updated. This will allow us to use optimistic locking to prevent race conditions when updating user records, in the same way that we did with movies earlier in the book.

Execute the migration
```sql
migrate -path=./migrations -database=$GREENLIGHT_DB_DSN up
```
