# Environment Checker Function Image

This images compares the payload info with the header.

## Requirements

- IronFunctions API

## Development

### Setup function file

```

fn init $USERNAME/func-checker
```

### Building image locally

```
fn build
```

### Publishing to DockerHub

```
fn push
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
```

### Running with IronFunctions

With this command we are going to create an application with name `checker`.

```
fn apps create checker
```

Now, we can create our route

```
fn routes create checker /check --config "TEST=1"
```

#### Testing function

Now that we created our IronFunction route, let's test our new route

```
echo '{ "env_vars": { "test": "1" } }' | fn call checker /check
```