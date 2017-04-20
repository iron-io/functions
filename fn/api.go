package main

import (
	"log"
	"net/url"
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	fnclient "github.com/iron-io/functions_go/client"
)

func apiURL() *url.URL {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		log.Fatalln("Couldn't parse API URL:", err)
	}

	return u
}

func host() string {
	return apiURL().Host
}

func scheme() string {
	return apiURL().Scheme
}

func apiClient() *fnclient.Functions {
	url := apiURL()

	transport := httptransport.New(url.Host, "/v1", []string{url.Scheme})

	if os.Getenv("IRON_TOKEN") != "" {
		transport.DefaultAuthentication = httptransport.BearerToken(os.Getenv("IRON_TOKEN"))
	}

	// create the API client, with the transport
	client := fnclient.New(transport, strfmt.Default)

	return client
}
