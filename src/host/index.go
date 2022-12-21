package host

import (
	"crypto/x509"
	"errors"
	"net/http"
	"strings"
)

var client = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func GetHostSSLInfo(host string) (*x509.Certificate, error) {
	if !strings.HasPrefix(host, "https") {
		return nil, errors.New("use https scheme")
	}
	request, err := http.NewRequest("get", host, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.TLS != nil {
		certificates := response.TLS.PeerCertificates
		if len(certificates) > 0 {
			return certificates[0], nil
		}
	}
	return nil, errors.New("TLS connection fail")
}
