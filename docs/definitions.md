# API Definitions

## App

Represents a unique `app` in the API and is identified by the property `name`

### App Example

```json
{
    "name": "myapp",
    "myconfig": "config"
}
```

### Properties

#### name (string)

`Name` is a property that references an unique app. 

#### config (object)

`Config` is a set of configurations that will be passed to all functions inside this app. 

## Route

Represents a unique `route` in the API and is identified by the property `path` and `app`.

Every `route` belongs to an `app`.

### Route Example
```json
{
    "path": "/hello",
    "image": "iron/hello",
    "type": "sync",
    "memory": 128,
    "config": {
        "key": "value",
        "key2": "value2",
        "keyN": "valueN",
    },
    "headers": {
        "content-type": [
            "text/plain"
        ]
    }
}
```

### Properties

#### path (string)

`Path` is the unique representation from a route inside a app. The combination of the `Path` and the `AppName` referers to a unique route.

Every `Path` must start with a `/` (dash)

#### image (string)

`Image` is the name or registry URL that references to a valid container image located locally or in a remote registry (if provided any registry address).

If no registry is provided and image is not available locally the API will try pull it from a default public registry.

#### type (string)

Options: `sync` and `async`

`Type` is defines how the function will be executed. If type is `sync` the request will be hold until the result is ready and flushed.

In `async` functions the request will be ended with a `call_id` and the function will be executed in the background.

#### memory (number)

`Memory` defines the amount of memory (in megabytes) required to run this function.

#### config (object of string values)

`Config` is a set of configurations that will be passed to the function.

If any `route` configuration value conflicts with an `app` configuration, the `route` value will be used.

#### headers (object of array of string)

`Header` is a set of headers that will be sent in the function execution response. The header value is an array of strings.

#### format (string)

`Format` defines if the function is running or not in `hot container` mode.

To define the function execution as `hot container` you set it as one of the following formats:

- `"http"`

### 'Hot Container' Only Properties

This properties are only used if the function is in `hot container` mode

#### max_concurrency (string)

This property defines the maximum amount of concurrent hot containers instances the function should have (per IronFunction node). 