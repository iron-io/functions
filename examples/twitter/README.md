# Twitter Function Image

This function exemplifies an authentication in Twitter API and get latest tweets of an account.

## Requirements

- IronFunctions API
- Configure a [Twitter App](https://apps.twitter.com/) and [configure Customer Access and Access Token](https://dev.twitter.com/oauth/overview/application-owner-access-tokens).

## Development

### Setup function file

```
fn init USERNAME/twitter
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

CUSTOMER_KEY="XXXXXX"
CUSTOMER_SECRET="XXXXXX"
ACCESS_TOKEN="XXXXXX"
ACCESS_SECRET="XXXXXX"
```

### Running with IronFunctions

With this command we are going to create an application with name `twitter`.

```
fn apps create twitter
fn apps config set twitter CUSTOMER_KEY $CUSTOMER_KEY
fn apps config set twitter CUSTOMER_SECRET $CUSTOMER_SECRET
fn apps config set twitter ACCESS_TOKEN $ACCESS_TOKEN
fn apps config set twitter ACCESS_SECRET $ACCESS_SECRET
```

Now, we can create our route

```
fn routes create tweeter /tweets
```

#### Testing function

Now that we created our IronFunction route, let's test our new route

```
curl -X POST --data '{"username": "getiron"}' http://$FUNCAPI/r/twitter/tweets
```