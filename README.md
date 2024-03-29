# Overview

It's an Clean Architecture Skeleton project based on Chi framework.
Our aim is reducing development time on default features that you can meet very often when your work on API.
There is a useful set of tools that described below. Feel free to contribute!


## What's inside:
- Authentication with Oauth 2.0
- CRUD API
- Migrations
- Request validation
- Swagger docs
- Environment configuration
- Docker development environment


## Usage

1. Copy .env.dist to .env and set the environment variables. In the .env file set these variables:
2. Browse to {HTTP_HOST}:{HTTP_PORT}/swagger/index.html. You will see Swagger 2.0 API documents.


## Directories

1. **main.go**: contains the application's main entry point(s) or command-line interfaces (CLIs). Each subdirectory represents a different executable within the project
2. **/internal**: houses the internal components of your application that are not intended to be imported by external projects. This directory typically contains packages/modules related to business logic, domain models, repositories, services, and configuration.
3. **/internal/app**: this section may include any initialization code that needs to be executed before the application starts. For example, setting up configuration, connecting to databases, or initializing logging.
4. **/internal/cache**: directory allows for the separation of caching concerns from other parts of the application, promoting modularity and maintainability. By isolating caching-related code, it becomes easier to manage and test caching functionality independently. However, the specific directory structure and organization may vary based on the project's needs and preferences.
5. **/internal/config**: holds the configuration-related code and files. It includes the logic to read and parse configuration files, environment variables, or other sources of configuration data. It provides a centralized way to manage and access application configuration throughout the codebase.
6. **/internal/domain**: directory, you separate the core business logic from infrastructure-specific or framework-specific code. This separation helps keep your code clean, maintainable, and easier to test. It also allows for better reusability and modularity, as the domain layer can be used independently of the specific infrastructure or framework being used.
7. **/internal/handler**: contains the HTTP or RPC handlers for the application. These handlers are responsible for receiving incoming requests, parsing them, invoking the necessary business logic, and returning the appropriate responses. Each handler typically corresponds to a specific endpoint or operation in the application's API.
8. **/internal/repository**: contains the implementation of data access and persistence logic. It provides an abstraction over the data storage layer, allowing the application to interact with databases, or other external systems. Repositories handle the CRUD operations and data querying required by the application.
9. **/internal/service**: contains the implementation of the application's business logic. It encapsulates the core functionality of the application and provides high-level operations that the handlers can use to accomplish specific tasks. Services interact with data repositories, external APIs, or other dependencies to fulfill the application's requirements.
10. **/migrations/{store}**: contains database migration scripts, which are used to manage database schema changes over time.
11. **/pkg**: contains packages that can be imported and used by external projects. These packages are typically utilities, libraries, or modules that have potential for reuse across different projects.


## Libraries

1. Router: https://github.com/go-chi/chi
2. Migrations: https://github.com/golang-migrate/migrate
3. Swagger: https://github.com/swaggo/swag


# Migrations: PostgreSQL tutorial for beginners

## Create/configure database

For the purpose of this tutorial let's create PostgreSQL database called `example`.
Our user here is `postgres`, password `password`, and host is `localhost`.
```
psql -h localhost -U postgres -w -c "create database example;"
```
When using Migrate CLI we need to pass to database URL. Let's export it to a variable for convenience:
```
export POSTGRESQL_URL='postgres://postgres:password@localhost:5432/example?sslmode=disable'
```
`sslmode=disable` means that the connection with our database will not be encrypted. Enabling it is left as an exercise.


## Create migrations
Let's create table called `users`:
```
migrate create -ext sql -dir db/migrations -seq create_users_table
```
If there were no errors, we should have two files available under `db/migrations` folder:
- 000001_create_users_table.down.sql
- 000001_create_users_table.up.sql

Note the `sql` extension that we provided.

In the `.up.sql` file let's create the table:
```
CREATE TABLE IF NOT EXISTS users(
   user_id serial PRIMARY KEY,
   username VARCHAR (50) UNIQUE NOT NULL,
   password VARCHAR (50) NOT NULL,
   email VARCHAR (300) UNIQUE NOT NULL
);
```
And in the `.down.sql` let's delete it:
```
DROP TABLE IF EXISTS users;
```
By adding `IF EXISTS/IF NOT EXISTS` we are making migrations idempotent - you can read more about idempotency in [getting started](../../GETTING_STARTED.md#create-migrations)

## Run migrations
```
migrate -database ${POSTGRESQL_URL} -path db/migrations up
```
Let's check if the table was created properly by running `psql example -c "\d users"`.
The output you are supposed to see:
```
                                    Table "public.users"
  Column  |          Type          |                        Modifiers                        
----------+------------------------+---------------------------------------------------------
 user_id  | integer                | not null default nextval('users_user_id_seq'::regclass)
 username | character varying(50)  | not null
 password | character varying(50)  | not null
 email    | character varying(300) | not null
Indexes:
    "users_pkey" PRIMARY KEY, btree (user_id)
    "users_email_key" UNIQUE CONSTRAINT, btree (email)
    "users_username_key" UNIQUE CONSTRAINT, btree (username)
```
Great! Now let's check if running reverse migration also works:
```
migrate -database ${POSTGRESQL_URL} -path db/migrations down
```
Make sure to check if your database changed as expected in this case as well.

## Database transactions

To show database transactions usage, let's create another set of migrations by running:
```
migrate create -ext sql -dir db/migrations -seq add_mood_to_users
```
Again, it should create for us two migrations files:
- 000002_add_mood_to_users.down.sql
- 000002_add_mood_to_users.up.sql

In Postgres, when we want our queries to be done in a transaction, we need to wrap it with `BEGIN` and `COMMIT` commands.
In our example, we are going to add a column to our database that can only accept enumerable values or NULL.
Migration up:
```
BEGIN;

CREATE TYPE enum_mood AS ENUM (
	'happy',
	'sad',
	'neutral'
);
ALTER TABLE users ADD COLUMN mood enum_mood;

COMMIT;
```
Migration down:
```
BEGIN;

ALTER TABLE users DROP COLUMN mood;
DROP TYPE enum_mood;

COMMIT;
```


## Optional: Run migrations within your Go app
Here is a very simple app running migrations for the above configuration:
```
import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	m, err := migrate.New(
		"file://db/migrations",
		"postgres://postgres:postgres@localhost:5432/example?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		log.Fatal(err)
	}
}
```


# Swagger: HTTP tutorial for beginners

1. Add comments to your API source code, See [Declarative Comments Format](#declarative-comments-format).

2. Download swag by using:
```sh
go install github.com/swaggo/swag/cmd/swag@latest
```
To build from source you need [Go](https://golang.org/dl/) (1.17 or newer).

Or download a pre-compiled binary from the [release page](https://github.com/swaggo/swag/releases).

3. Run `swag init` in the project's root folder which contains the `main.go` file. This will parse your comments and generate the required files (`docs` folder and `docs/docs.go`).
```sh
swag init
```

  Make sure to import the generated `docs/docs.go` so that your specific configuration gets `init`'ed. If your General API annotations do not live in `main.go`, you can let swag know with `-g` flag.
  ```sh
  swag init -g internal/handler/handler.go
  ```

4. (optional) Use `swag fmt` format the SWAG comment. (Please upgrade to the latest version)

  ```sh
  swag fmt
  ```