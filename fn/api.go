package main

import (
	"os"

	"crypto/tls"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	fnclient "github.com/iron-io/functions_go/client"
	"log"
	"net/http"
	"net/url"
)

func host() string {
	u, err := url.Parse(API_URL)
	if err != nil {
		log.Fatalln("Couldn't parse API URL:", err)
	}

	return u.Host
}

func apiClient() *fnclient.Functions {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: SSL_SKIP_VERIFY},
	}
	cl := &http.Client{Transport: tr}

	transport := httptransport.NewWithClient(host(), API_VERSION, []string{"http", "https"}, cl)
	if os.Getenv("IRON_TOKEN") != "" {
		transport.DefaultAuthentication = httptransport.BearerToken(os.Getenv("IRON_TOKEN"))
	}

	// create the API client, with the transport
	client := fnclient.New(transport, strfmt.Default)

	return client
}
