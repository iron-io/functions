## Quick Example for a Python Function (4 minutes)

This example will show you how to test and deploy Python code to IronFunctions.

### 1. Setup function file

```sh
fn init USERNAME/hello
```

### 2. Build:

```sh
# build the function
fn build
# test it
cat hello.payload.json | fn run
# push it to Docker Hub
fn push
# Create a route to this function on IronFunctions
fn routes create pythonapp /hello
```

`-v` is optional, but it allows you to see how this function is being built.

### 3. Queue jobs for your function

Now you can start jobs on your function. Let's quickly queue up a job to try it out.

```sh
cat hello.payload.json | fn call pythonapp /hello
```

Here's a curl example to show how easy it is to do in any language:

```sh
curl -H "Content-Type: application/json" -X POST -d @hello.payload.json http://localhost:8080/r/pythonapp/hello
```