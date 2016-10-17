Async Image Processing Function
================================

An image processing Docker image

## Setup your Function service

First off, you need to setup [Functions](https://github.com/iron-io/functions#quickstart).

Second, copy `payload_example.json` to `payload.json` and fill in your AWS credentials.

Now you can get cracking!

## Usage

**NOTE**: Replace `YOUR_USERNAME` everywhere below with your Docker Hub username.

The payload for this function defines the image operations you'd like to perform and where to store the results. See `payload_example.json` for an example.

## Test / build

### 1. Build Docker image

```sh
# SET BELOW TO YOUR DOCKER HUB USERNAME
USERNAME=YOUR_DOCKER_HUB_USERNAME

# build it
docker build -t $USERNAME/image_processor .
```

### 2. Test Docker image with a single image

```sh
cat payload.json | docker run --rm -it $USERNAME/image_processor:0.0.1
```

### 3. Push to Docker Hub

```sh
docker push $USERNAME/image_processor
```

### 4. Modifying this function

After modifying this function you can run this commands to bump your function version

```
docker run --rm -v "$PWD":/app treeder/bump patch
docker tag $USERNAME/image_processor:latest $USERNAME/image_processor:`cat VERSION`
```

## Run it in IronFunctions

Now that we have our function built as an image and it's up on Docker Hub,
we can start using that to process massive amounts of images.  

Start your IronFunctions API and save this environment var:

```sh
# If you are running it locally the endpoint might be http://127.0.0.1:8080
ENDPOINT=ADDRESS_TO_FUNCTIONS_ENDPOINT
```

### 1. Creating your new application

```sh
# Let's call it 'image'
curl -H "Content-Type: application/json" -X POST -d '{
    "app": { "name":"image" }
}' $ENDPOINT/v1/apps
```

### 2. Now we can define a route to trigger our new function

```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "route": {
        "path":"/process",
        "image":"'$USERNAME'/image_processor:latest"
    }
}' $ENDPOINT/v1/apps/image/routes
```

### 3. Great, now our function is ready. Let's test it.

```sh
# Let's send our payload.json content to our new route
cat payload.json | curl -X POST $ENDPOINT/r/image/process -H "Content-Type: application/json" -d @-
```