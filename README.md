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


# BUILD DOCKER IMAGE for Construction API

## prerequisite
- docker version 19.+
- AWS CLI v2 
  - [Linux](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-linux.html)
  - [MacOS](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-mac.html)
  - [Windows](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-windows.html)

## Configure AWS CLI
use command below to configure AWS CLI 

```$ aws configure```

Input the AWS Access ID and Secret.

## DEPLOYMENT

Before we start for building the docker image, you need to know that we are using 3 different of environment
- dev
- stage
- prod

We are using this 3 env for controlling the version. example: 
- dbo/core-services:1.0.0-dev
- dbo/core-services:1.0.0-stage
- dbo/core-services:1.0.0-prod


### Build dbo-core image

1. build the :latest release of the image first.

```$ docker build --no-cache -t dbo/core-services:latest .```

Make sure that you are using the same repository name with content above (`dbo/core-services`).

2. give a tag for this image

```$ docker tag dbo/core-services:latest 011876906689.dkr.ecr.us-west-2.amazonaws.com/dbo/core-services:1.x.x-env```

Make sure you are using sequence for versioning and you must know what is the latest version of this image. Example: `dbo/core-services:1.0.1-stage` need change to `dbo/core-services:1.0.2-stage`. The ECR service will reject the image if the version tag **is exists** on iamge repository.

3. push the new image to ECR repository.

```$ docker push 011876906689.dkr.ecr.us-west-2.amazonaws.com/dbo/core-services:1.0.2```

Please make sure that you are logged in to AWS ECR. Use this command to login:

```$ aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 011876906689.dkr.ecr.us-west-2.amazonaws.com```

This command will read from aws credentials that you set before (aws configure).

### Deploy To Server

1. Login to server
2. Find and edit the `.env` file in `deploy/` directory. You need to edit the content of `.env` file: 

```
API_IMAGE=011876906689.dkr.ecr.us-west-2.amazonaws.com/dbo/core-services:1.0.1-stage
``` 
change the version to

```
API_IMAGE=011876906689.dkr.ecr.us-west-2.amazonaws.com/dbo/core-services:1.0.2-stage
```
and save it.

3. Login to AWS ECR using this command below:

```$ aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 011876906689.dkr.ecr.us-west-2.amazonaws.com```

4. Pull the image and deploy using docker-compose:

```$ docker-compose up -d```


## IMPORTANT !!!
**Please make sure that you edit the version table above.* 

I'ts easy for us to know what is the latest version of the docker image.




