# Quick Example for a Go Function (3 minutes)

This example will show you how to test and deploy Go (Golang) code to IronFunctions.

```sh
# create your func.yaml file
fn init <YOUR_DOCKERHUB_USERNAME>/hello
# build the function
fn build
# test it
cat hello.payload.json | fn run
# Create a route to this function on IronFunctions
fn deploy myapp
```

Now you can call your function on IronFunctions:

```sh
curl -H "Content-Type: application/json" -X POST -d @hello.payload.json http://localhost:8080/r/myapp/hello
```

## Dependencies

Be sure you're dependencies are in the `vendor/` directory and that's it.

