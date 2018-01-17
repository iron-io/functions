# Error Function Image

This images compares the payload info with the header.

```
# Create your func.yaml file
fn init <YOUR_DOCKERHUB_USERNAME>/func-error
# Build the function
fn build
# Test it
echo '{"input": "yoooo"}' | fn run
# Push it to Docker Hub
fn push
# Create routes to this function on IronFunctions
fn apps create <YOUR_APP>
fn routes create <YOUR_APP> /error
```

## Running it on IronFunctions

### Let's define some environment variables

```
# Set your Function server address
# Eg. 
# FUNCAPI=127.0.0.1:8080
# APPNAME=error
FUNCAPI=YOUR_FUNCTIONS_ADDRESS
APPNAME=YOUR_APP
```

### Testing function

Now that we created our IronFunction route, let's test our new route

```
curl -X POST --data '{"input": "yoooo"}' http://$FUNCAPI/r/$APPNAME/error
```
