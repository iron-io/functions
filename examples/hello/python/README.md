# Quick Example for a Python Function (4 minutes)

This example will show you how to test and deploy Go (Golang) code to IronFunctions.

```sh
# create your func.yaml file
fn init <YOUR_DOCKERHUB_USERNAME>/hello
# build/get dependencies
docker run --rm -v %s:/worker -w /worker iron/python:2-dev pip install -t packages -r requirements.txt
# build the function
fn build
# test it
cat hello.payload.json | fn run
# push it to Docker Hub
fn push
# Create a route to this function on IronFunctions
fn routes create pythonapp /hello
```

Now you can use your functions at the URL http://localhost:8080/r/pythonapp/hello. Let's quickly call our function to try it out.

```sh
cat hello.payload.json | fn call pythonapp /hello
```

Here's a curl example to show how easy it is to do in any language:

```sh
curl -H "Content-Type: application/json" -X POST -d @hello.payload.json http://localhost:8080/r/pythonapp/hello
```
