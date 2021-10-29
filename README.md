# CONSTRUCTION PROJECT API 

## Introduction

- Minimum Go Requirement 1.13

``` ~ https://github.com/fendys01/Construction-Project.git```


- COPY config.json.example TO config.json

``` ~ cp -r config.json.example config.json ```

- GENERATE APP KEY 

Please visit http://www.sha1-online.com to generate a new key and add the generated key to config.json
``` 
{
    "app": {
        ....
        "key": "paste the generated key here!"
        ....
    },
....
```

## install all dependencies

```~ go mod download```

## Migrations

Install the migration tool and follow this [link](https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md) for the instructions.


**Create Migration**

please follow this [link](https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md)

example: 

``` migrate create -ext sql -dir resources/migrations -seq create_[table name]_table ```

**Execute Migration**

add an env variable for psql connection to your system:

```export POSTGRESQL_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable&search_path=public'```

and run the command to execute migrations:

``` migrate -database ${POSTGRESQL_URL} -path db/migrations up ```


## Available Channel

There is 3 channel that will be used on this api:
- cust_mobile_app
- cms
- tc_mobile_app