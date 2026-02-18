package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func LoadTLSConfig(certPath string) (*tls.Config, error) {
	if certPath == "" {
		return nil, nil
	}

	caCert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения сертификата: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("ошибка парсинга сертификата")
	}

	return &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}, nil
}
