package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Person Structure will bind the body accepted by this function
type Person struct {
	Name string `json:"Name"`
}

// handler function deals with the incoming request in http api style
func handler(req *http.Request, res *http.Response) (string, error) {
	// If cant read the http request
	// Reading the body
	p := &Person{Name: "World"}
	err := json.NewDecoder(req.Body).Decode(p)
	// If the body is not correct
	if err != nil || len(p.Name) == 0 {
		res.StatusCode = 400
		res.Status = http.StatusText(res.StatusCode)
		http.StatusText(res.StatusCode)
		return "Invalid Body ", err
	}
	return fmt.Sprintf("Hello %s", p.Name), err
}

// main function can be left unchanged in simple applications. You should be able to make changes to
func main() {
	for {
		// Initializing response structure and the Buffer it will use
		res := http.Response{
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			StatusCode: 200,
			Status:     "OK",
		}
		var buf bytes.Buffer

		// reading http request from stdin
		req, err := http.ReadRequest(bufio.NewReader(os.Stdin))
		if err != nil {
			res.StatusCode = 500
			res.Status = http.StatusText(res.StatusCode)
			fmt.Fprintln(&buf, err)
		} else {
			response, err := handler(req, &res)
			if err != nil {
				fmt.Fprintf(&buf, err.Error())
			} else {
				fmt.Fprintf(&buf, response)
			}
		}
		res.Body = ioutil.NopCloser(&buf)
		res.ContentLength = int64(buf.Len())
		res.Write(os.Stdout)
	}
}
