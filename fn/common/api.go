package common

import (
	"crypto/tls"
	"net/http"
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	fnclient "github.com/iron-io/functions_go/client"
)

func ApiClient() *fnclient.Functions {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: SSL_SKIP_VERIFY},
	}
	cl := &http.Client{Transport: tr}

	transport := httptransport.NewWithClient(HOST, API_VERSION, []string{SCHEME}, cl)
	if os.Getenv("IRON_TOKEN") != "" {
		transport.DefaultAuthentication = httptransport.BearerToken(os.Getenv("IRON_TOKEN"))
	}

	// create the API client, with the transport
	client := fnclient.New(transport, strfmt.Default)

	return client
}
