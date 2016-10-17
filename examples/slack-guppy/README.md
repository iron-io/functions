# Slack Command Function

## 1. Building the function image

```sh
USERNAME=YOUR_DOCKER_HUB_USERNAME
docker build -t $USERNAME/func-guppy:`cat VERSION` .
```

## 2. Testing the function

```sh
cat slack.payload | docker run --rm -i $USERNAME/func-guppy:`cat VERSION`
```

## 3. Running in IronFunctions

### Publish your function to DockerHub.

```sh
docker push $USERNAME/func-guppy:`cat VERSION`
```

### Creating our IronFunctions route to run this function

First let's define a environment var to our IronFunctions endpoint 

```sh
# If it's running locally might be http://127.0.0.1:8080
ENDPOINT=IRON_FUNCTIONS_ENDPOINT
```

Now let's create our function application, called `slack`

```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "app": { "name":"slack" }
}' $ENDPOINT/v1/apps
```

And a route to that function:

```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "route": {
        "path":"/guppy",
        "image":"'$USERNAME'/func-guppy:'`cat VERSION`'"
    }
}' $ENDPOINT/v1/apps/slack/routes
```

Now we can just run our function on the url `$ENDPOINT/r/slack/guppy`

```sh
cat slack.payload | curl -X POST $ENDPOINT/r/slack/guppy -H "Content-Type: application/json" -d @-
```

## 4. Modifying the function

When you modify this function locally, make sure to bump its version using the following commands:

```sh
docker run --rm -v "$PWD":/app treeder/bump patch
docker tag $USERNAME/func-guppy:latest $USERNAME/func-guppy:`cat VERSION`
```