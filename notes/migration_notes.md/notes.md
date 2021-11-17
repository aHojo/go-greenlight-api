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

## Migration tool  

We will use the migrate command line tool.  

[migrate](https://github.com/golang-migrate/migrate)

```bash
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

### migrations folder now  

```sql
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
 migrate create -seq -ext=.sql -dir=./migrations create_users_table
/home/ahojo/development/go/src/go-greenlight-api/migrations/000004_create_users_table.up.sql
/home/ahojo/development/go/src/go-greenlight-api/migrations/000004_create_users_table.down.sql
```

04 up

```sql
CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
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

1. The email column has the type citext (case-insensitive text). This type stores text data exactly as it is inputted — without changing the case in any way — but comparisons against the data are always case-insensitive... including lookups on associated indexes.

2. We’ve also got a `UNIQUE` constraint on the  `email` column. Combined with the `citext` type, this means that no two rows in the database can have the same `email` value — even if they have different cases. This essentially enforces a database-level business rule that no two users should exist with the same `email` address.  

3. The `password_hash` column has the type `bytea` (binary string). In this column we’ll store
a one-way hash of the user’s password generated using `bcrypt` — not the plaintext password

4. The activated column stores a `boolean` value to denote whether a user account is ‘active’ or not. We will set this to false by default when creating a new user, and require the user to confirm their email address before we set it to true.

5. We’ve also included a version number column, which we will increment each time a user record is updated. This will allow us to use optimistic locking to prevent race conditions when updating user records, in the same way that we did with movies earlier in the book.

Execute the migration

```sql
migrate -path=./migrations -database=$GREENLIGHT_DB_DSN up
```

05 Tokens Table

This will be used to identify and activate a user.  

```sql
migrate create -seq -ext .sql -dir ./migrations create_tokens_table
/home/ahojo/development/go/src/go-greenlight-api/migrations/000005_create_tokens_table.up.sql
/home/ahojo/development/go/src/go-greenlight-api/migrations/000005_create_tokens_table.down.sql
```

up migration

```sql
CREATE TABLE IF NOT EXISTS tokens (
  hash bytea PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  expiry timestamp(0) with time zone NOT NULL,
  scope text NOT NULL
);
```

down migration

```sql
DROP TABLE IF EXISTS tokens;
```

- The hash column will contain a SHA-256 hash of the activation token. It’s important to
emphasize that we will only store a hash of the activation token in our database — not
the activation token itself.
- The `user_id` column will contain the ID of the user associated with the token. We use the REFERENCES user syntax to create a foreign key constraint against the primary key of our users table, which ensures that any value in the `user_id` column has a corresponding id entry in our users table.
*We also use the ON `DELETE CASCADE` syntax to instruct PostgreSQL to automatically delete all records for a user in our tokens table when the parent record in the users table is deleted.*
- The expiry column will contain the time that we consider a token to be ‘expired’ and no
longer valid.  

- Lastly, the scope column will denote what purpose the token can be used for. Later in the book we’ll also need to create and store authentication tokens, and most of the code and storage requirements for these is exactly the same as for our activation tokens. So instead of creating separate tables (and the code to interact with them), we’ll store them in one table with a value in the scope column to restrict the purpose that the token can be used for.

Do the migrations  

```sql
migrate -path=./migrations -database=$GREENLIGHT_DB_DSN up
```

## Permissions and many-to-many relationship  

UP
`migrate create -seq -ext .sql -dir ./migrations add_permissions`

PERFORM THE MIGRATION
`$ migrate -path ./migrations -database $GREENLIGHT_DB_DSN up`

- The PRIMARY KEY (user_id, permission_id) line sets a composite primary key on our users_permissions table, where the primary key is made up of both the users_id and permission_id columns. Setting this as the primary key essentially means that the same user/permission combination can only appear once in the table and cannot be duplicated.

- When creating the users_permissions table we use the REFERENCES user syntax to create a foreign key constraint against the primary key of our users table, which ensures that any value in the user_id column has a corresponding entry in our users table. And likewise, we use the REFERENCES permissions syntax to ensure that the permission_id column has a corresponding entry in the permissions table.

UP

```sql
CREATE TABLE IF NOT EXISTS permissions (
  id bigserial PRIMARY KEY,
  code text NOT NULL
);

CREATE TABLE IF NOT EXISTS users_permissions (
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE,
  PRIMARY KEY (user_id, permission_id)
);

-- Add the two permissions to the table.
INSERT INTO permissions (code)
VALUES
('movies:read'),
('movies:write');
```

DOWN

```sql
DROP TABLE IF EXISTS users_permissions;
DROP TABLE IF EXISTS permissions;
```

## Make changes in the DB to give users permissions  

```sql
-- Set the activated field for alice@example.com to true.
UPDATE users SET activated = true WHERE email = 'alice@example.com';
-- Give all users the 'movies:read' permission
INSERT INTO users_permissions
SELECT id, (SELECT id FROM permissions WHERE code = 'movies:read') FROM users;
-- Give faith@example.com the 'movies:write' permission
INSERT INTO users_permissions
VALUES (
(SELECT id FROM users WHERE email = 'faith@example.com'),
(SELECT id FROM permissions WHERE code = 'movies:write')
);
-- List all activated users and their permissions.
SELECT email, array_agg(permissions.code) as permissions
FROM permissions
INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
INNER JOIN users ON users_permissions.user_id = users.id
WHERE users.activated = true
GROUP BY email;
```

Note: In that final SQL query, we’re using the aggregation function array_agg() and a
GROUP BY clause to output the permissions associated with each email address as an
array.
