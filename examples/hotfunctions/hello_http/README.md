# Hello World HTTP Function

This function is a HTTP style implementation of the basic hello world.

It is implemented in a way that a handler function will deal with default http requests and responses, similar to how it would work in any other framework.


```go
    func handler(req *http.Request, res *http.Response) (string, error) {
        
        decoder := json.NewDecoder(req.Body) // decode data from body
        
        ...

        return fmt.Sprintf("Content to return to client"), err
    }
```