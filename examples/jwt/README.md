# Quick Example for a Authentication

This example will show you how to test and deploy secure function to IronFunctions.

```sh
# create your func.yaml file
fn init <YOUR_DOCKERHUB_USERNAME>/jwt

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

Now you can call your function on IronFunctions:

```sh
# Get token for authentication
fn routes token myapp /jwt
#The token expiration time is 1 hour by default. You can specify an expiration time with such arguments.
#e.g. The token expires by 500 seconds.
fn routes token myapp /jwt 500

#You'll get token like this
# {
#        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MDgwNTcwNTEsImlhdCI6MTUwODA1MzQ1MX0.3c_xUaleCdHy_fdU9zFB50j3hqwYWgPZ-EkTXV3VWag"
# }

#Access to your app with token
curl  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MDgwNTcwNTEsImlhdCI6MTUwODA1MzQ1MX0.3c_xUaleCdHy_fdU9zFB50j3hqwYWgPZ-EkTXV3VWag' http://localhost:8080/r/myapp/jwt

# or use
fn call myapp /jwt

```

## Dependencies

Be sure you're dependencies are in the `vendor/` directory and that's it.

