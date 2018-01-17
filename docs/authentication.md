# Authentication

Iron Functions API supports two levels of Authentication in two seperate scopes,
using [JWT](https://jwt.io/) tokens for authentication.

* Service level authentication
* Route level authentication.

## Service level authentication

This authenticates all requests made to the server from any client. Service level authentication
is set by the `JWT_AUTH_KEY` environment variable. If you set with variable
within the server context, the server will look for a valid JWT token in every request. For use with the
`fn` tool, you also need to set this environment variable within `fn` tool context while calling a command.

## Route level authentication

Route level authentication is applied whenever a function call made to a specific route. You can check
[Quick Example for JWT Authentication](../examples/jwt/README.md) for an example.
