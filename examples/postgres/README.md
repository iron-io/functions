# Postgres INSERT/SELECT Function Image

This function executes an INSERT or SELECT against a table in a given postgres server.

## Requirements

- Postgres Server
- IronFunctions API

## Development

### Building image locally

```
# SET BELOW TO YOUR DOCKER HUB USERNAME
USERNAME=YOUR_DOCKER_HUB_USERNAME

# build it
./build.sh
```

### Publishing to DockerHub

```
# tagging
docker run --rm -v "$PWD":/app treeder/bump patch
docker tag $USERNAME/func-postgres:latest $USERNAME/func-postgres:`cat VERSION`

# pushing to docker hub
docker push $USERNAME/func-postgres
```

### Testing image

```
./test.sh
```

## Running it on IronFunctions

### Let's define some environment variables

```
# Set your Function server address
# Eg. 127.0.0.1:8080
FUNCAPI=YOUR_FUNCTIONS_ADDRESS

# Set your Postgres server address
# Eg. postgres:5432
POSTGRES=YOUR_POSTGRES_ADDRESS

# Set your table name
# Eg. people
TABLE=YOUR_TABLE_NAME
```

### Running with IronFunctions

This command creates an application named `postgres`.

```
curl -X POST --data '{
    "app": {
        "name": "postgres",
        "config": {
            "server": "'$POSTGRES'"
        }
    }
}' http://$FUNCAPI/v1/apps
```

Now, we can create our routes.

#### Route for insert value

```
curl -X POST --data '{
    "route": {
        "image": "'$USERNAME'/func-postgres",
        "path": "/$TABLE/insert",
        "config": {
            "command": "INSERT",
            "table": "$TABLE"
        }
    }
}' http://$FUNCAPI/v1/apps/postgres/routes
```

#### Route for select value

```
curl -X POST --data '{
    "route": {
        "image": "'$USERNAME'/func-postgres",
        "path": "/$TABLE/select",
        "config": {
            "command": "SELECT",
            "table": "$TABLE"
        }
    }
}' http://$FUNCAPI/v1/apps/postgres/routes
```

#### Testing function

Now that we created our IronFunction routes, let's test them.

```
curl -X POST --data '{"first": "John", "last": "Smith"}' http://$FUNCAPI/r/postgres/people/insert
// "OK"
curl -X POST --data '{"first": "Bob", "last": "Smith"}' http://$FUNCAPI/r/postgres/people/insert
// "OK"
curl -X POST --data '{"last": "Smith"}' http://$FUNCAPI/r/postgres/people/select
// [{"first": "John", "last": "Smith"}, {"first": "Bob", "last": "Smith"}]
curl -X POST --data '{"first": "Bob"}' http://$FUNCAPI/r/postgres/people/select
// [{"first": "Bob", "last": "Smith"}]
```