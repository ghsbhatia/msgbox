package svcclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	kithttp "github.com/go-kit/kit/transport/http"
)

// Abstraction for HTTP Service Access for json content
type HttpServiceClient interface {
	// Send Http Get Request to given endpoint and unmarshal the Response
	// to given type.
	Get(string, interface{}) error
}

// Get a new instance of HttpServiceClient
func NewHttpClient() HttpServiceClient {
	return &httpserviceClient{}
}

// HttpServiceClient implementation
type httpserviceClient struct {
}

func (s *httpserviceClient) Get(svcurl string, v interface{}) error {
	encode := func(context.Context, *http.Request, interface{}) error {
		return nil
	}

	decode := func(_ context.Context, r *http.Response) (interface{}, error) {
		if r.StatusCode != http.StatusOK {
			return nil, errors.New(fmt.Sprintf("%d", r.StatusCode))
		}
		body, err := ioutil.ReadAll(r.Body)
		return body, err
	}

	endpoint, _ := url.Parse(svcurl)
	client := kithttp.NewClient("GET", endpoint, encode, decode)

	res, err := client.Endpoint()(context.Background(), struct{}{})
	if err != nil {
		return err
	}

	bytearray, _ := res.([]byte)
	err = json.Unmarshal(bytearray, v)

	return err
}
