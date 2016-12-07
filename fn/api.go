package main

import (
	"log"
	"net/url"
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	fnclient "github.com/iron-io/functions_go/client"
)

func host() string {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		log.Fatalln("Couldn't parse API URL:", err)
	}
	return apiURL.Host
}

func apiClient() *fnclient.Functions {
	transport := httptransport.New(getHost(), "/v1", nil)

	// create the API client, with the transport
	client := fnclient.New(transport, strfmt.Default)
	if os.Getenv("IRON_TOKEN") {
		client.DefaultAuthentication = fnclient.BearerToken(os.Getenv("IRON_TOKEN"))
	}
	return client
}
