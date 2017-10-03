# Environment Checker Function Image

This images compares the payload info with the header.

```
# Create your func.yaml file
fn init <YOUR_DOCKERHUB_USERNAME>/func-checker
# Build the function
fn build
# Test it
echo '{"env_vars": {"HEADER_FOO": "bar"}}' | ./fn run -e FOO=bar
# Push it to Docker Hub
fn push
# Create routes to this function on IronFunctions
fn apps create <YOUR_APP>
fn routes create <YOUR_APP> /check --config TEST=1
```

## Running it on IronFunctions

### Let's define some environment variables

```
# Set your Function server address
# Eg. 
# FUNCAPI=127.0.0.1:8080
# APPNAME=checker
FUNCAPI=YOUR_FUNCTIONS_ADDRESS
APPNAME=YOUR_APP
```

### Testing function

Now that we created our IronFunction route, let's test our new route

```
curl -X POST --data '{ "env_vars": { "TEST": "1" } }' http://$FUNCAPI/r/$APPNAME/check
```

