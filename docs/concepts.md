# IronFunctions Core Concepts

## Functions

**Function** is the unit responsible for handling a specific part of your service. Is usually formed by a small code that receives a payload, does its work and can return the result.

You can [configure a function](#Link-to-the-function-file-doc) adding a file called `func.yaml` in its source directory and setup how that function needs to run.

A function will run a task inside a container when triggered by a [route](#Route).

### Function Structure

```
name            string                   (username/function_name)
memory          number                   (max amount of memory used by the function)
headers         object{key => array}     (returned HTTP headers)
type            string (sync/async)      (how the function will be running)
format          string (http/-)          (how the function will receive the payload)
max_concurrency number                   (max concurrent containers can run per node)   
timeout         number                   (max amount of second can this function be kept alive)
config          object{key => value}     (configuration that will be passed to the function container)
```

## Routes

**Route** is the unit of the **IronFunctions** API that is responsible for triggering the function execution.

The route is defined by a `Path` (that will trigger the function execution), the `Function` that will be executed and `Tags` to help you manage all your functions.

### Route Structure

```
tags           object{key => value}     
path           string                   (path to trigger the function)
function       Function{}               (function that will be triggered)
```

## Apps

**App** is a collection of functions grouped by a tag name, in this case, by the tag `app`.

You can easily deploy multiple functions inside the same `App` creating a `app.yaml` file in the parent directory of your functions directory.

### App.yaml Structure

```
routes:
  /name-of-the-app/myRoute: {path to the function directory}
  /name-of-the-app/myOtherRoute:
    function: myOtherRoute/
    tag:
      private: true
```

In this case when you run `fn deploy`, the fn tool will create/update all routes with the configured **function** and **path** and tag all of them with `app = name-of-the-app`