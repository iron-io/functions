package common

import (
	"crypto/tls"
	"fmt"
	"net/http"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	f_common "github.com/iron-io/functions/common"
	fnclient "github.com/iron-io/functions_go/client"
)

func ApiClient() *fnclient.Functions {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: SSL_SKIP_VERIFY},
	}
	cl := &http.Client{Transport: tr}

	transport := httptransport.NewWithClient(HOST, API_VERSION, []string{SCHEME}, cl)

	if JWT_AUTH_KEY != "" {
		jwtToken, err := f_common.GetJwt(JWT_AUTH_KEY, 60*60)
		if err != nil {
			fmt.Println(fmt.Errorf("unexpected error: %s", err))
		} else {
			transport.DefaultAuthentication = httptransport.BearerToken(jwtToken)
		}
	}

	// create the API client, with the transport
	client := fnclient.New(transport, strfmt.Default)

	return client
}
