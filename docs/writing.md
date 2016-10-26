# Writing Functions

This will give you the basic overview of writing base level functions. You can also use higher level abstractions that make it easier such
as [lambda](lambda.md).

## Code

The most basic code layout in any language is as follows, this is pseudo code and is not meant to run.

```ruby
# Read and parse from STDIN
body = JSON.parse(STDIN)

# Do something
return_struct = doSomething(body)

# Respond
STDOUT.write(JSON.generate(return_struct))
```

## Packaging

Packaging is currently supported by creating a Docker image.

TODO
