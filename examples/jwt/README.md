# Quick Example for JWT Authentication

This example will show you how to test and deploy a function with JWT Authentication.

```sh
# create your func.yaml file
fn init <YOUR_DOCKERHUB_USERNAME>/<REPO NAME>

# Add 
# jwt_key: <Your JWT signing key>
# to your func.yml

# build the function
fn build
# test it
fn run
# push it to Docker Hub
fn push
# Create a route to this function on IronFunctions
fn routes create myapp /jwt


```

If you are going to add jwt authentication to an existing function,
you can simply add `jwt_key` to your func.yml, and update your route
using fn tool update command.

Now you can call your function on IronFunctions:

```sh
# Get token for authentication
fn routes token myapp /jwt
# The token expiration time is 1 hour by default. You can also specify the expiration time explicitly.
# Below example set the token expiration time at 500 seconds :
fn routes token myapp /jwt 500

# The response will include a token :
# {
#        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MDgwNTcwNTEsImlhdCI6MTUwODA1MzQ1MX0.3c_xUaleCdHy_fdU9zFB50j3hqwYWgPZ-EkTXV3VWag"
# }

# Now, you can access your app with a token :
curl  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MDgwNTcwNTEsImlhdCI6MTUwODA1MzQ1MX0.3c_xUaleCdHy_fdU9zFB50j3hqwYWgPZ-EkTXV3VWag' http://localhost:8080/r/myapp/jwt

# or use fn tool
# This will automatically generate a token and make function call :
fn routes call myapp /jwt

```

__important__: Please note that enabling Jwt authentication will require you to authenticate each time you try to call your function.
You won't be able to call your function without a token.

## Dependencies

Be sure your dependencies are in the `vendor/` directory.

