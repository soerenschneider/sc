package agent

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"go.uber.org/multierr"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/soerenschneider/sc-agent/pkg/api"
)

type HttpClientOption func(client *http.Client) error

func WithTlsClientCert(certFile, keyFile, caFile string) HttpClientOption {
	return func(client *http.Client) error {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}

		caCert, err := os.ReadFile(caFile)
		if err != nil {
			return err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS13,
		}

		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}

		return nil
	}
}

func BuildApp(server string, httpClientOpts ...HttpClientOption) (*ScAgentClient, error) {
	httpClient := retryablehttp.NewClient().HTTPClient

	var errs error
	for _, clientOpt := range httpClientOpts {
		if err := clientOpt(httpClient); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	if errs != nil {
		return nil, errs
	}

	apiOpts := []api.ClientOption{
		api.WithHTTPClient(httpClient),
	}

	client, err := api.NewClient(server, apiOpts...)
	if err != nil {
		return nil, err
	}

	return NewClient(client)
}
