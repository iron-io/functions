# Chained Workflow

You can use a chained workflow with IronFunctions, when creating a route, specify a targeting route to be its pipeline:

```sh
fnctl routes create otherapp /hasher iron/hasher
fnctl routes create otherapp /hello iron/hello --pipe /hasher
```

Or using cURL calls:
```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "route": {
        "path":"/hasher",
        "image":"iron/hasher"
    }
}' http://localhost:8080/v1/apps/myapp/routes

curl -H "Content-Type: application/json" -X POST -d '{
    "route": {
        "path":"/hello",
        "image":"iron/hello",
	"pipe":"/hasher"
    }
}' http://localhost:8080/v1/apps/myapp/routes
```

This configuration will take the output of `/hello` and feed it into `/hasher`, the response will always be returned to
the original caller. It is equivalent of doing:

```
fnctl routes run otherapp /hello | fnctl routes run otherapp /hasher
```

## Restrictions

### Loops

IronFuctions will return error in the following situations:
1 - Direct loops: routes who pipeline to themselves.
2 - Indirect loops: routes whose chain form loops, ie, some node points to some other known node.

### Timeout

If the work times out, it will not stop the chain. It will let the chain follow its course: either finishing or timing out too.