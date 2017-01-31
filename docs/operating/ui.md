# User Interface for IronFunctions

See Functions UI project https://github.com/iron-io/functions-ui

### Run Functions UI

```
docker run --rm -it --link functions:api -p 4000:4000 -e "API_URL=http://api:8080" iron/functions-ui
```
