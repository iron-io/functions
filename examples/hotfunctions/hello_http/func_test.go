package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// TestCorrectBody creates a request with a valid body to test the handler
func TestCorrectBody(t *testing.T) {
	body := strings.NewReader(`{"Name" : "Auyer"}`)
	req, err := http.NewRequest("POST", "http://api:8080/r/hot/hello", body)
	if err != nil {
		// handle err
	}

	req.Header.Set("Content-Type", "application/json")

	res := http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		StatusCode: 200,
		Status:     "OK",
	}

	response, err := handler(req, &res)
	if err != nil {
		t.Error(err.Error())
	} else if response != `Hello Auyer` {
		t.Error("Values do not match " + response)
	}
}

// TestCorrectBody creates a request with an invalid body to test the handler
func TestBadBody(t *testing.T) {
	body := strings.NewReader(`{"Name" : `)
	req, err := http.NewRequest("POST", "http://api:8080/r/primitives/multiply", body)
	if err != nil {
		// handle err
	}

	req.Header.Set("Content-Type", "application/json")

	res := http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		StatusCode: 200,
		Status:     "OK",
	}

	response, err := handler(req, &res)
	if err == nil {
		t.Error("Error not detected  error " + err.Error())
	} else if response != "Invalid Body " {
		t.Error("Returned value different than nil " + response)
	}
}

// TestCorrectBody creates a request with an empty body to test the handler
func TestEmptyBody(t *testing.T) {
	fmt.Print("Test")
	body := strings.NewReader("{}")
	req, err := http.NewRequest("POST", "http://api:8080/r/primitives/multiply", body)
	if err != nil {
		// handle err
	}

	req.Header.Set("Content-Type", "application/json")

	res := http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		StatusCode: 200,
		Status:     "OK",
	}

	response, err := handler(req, &res)
	if response != "Hello World" {
		t.Error("Returned value different than the Default " + response)
	}
}
