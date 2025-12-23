package connection

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"
)

type ConnectionManager struct {
	certPath          string
	connectionTimeout time.Duration
	maxRetries        int
	retryDelay        time.Duration
}

func NewConnectionManager(
	certPath string,
	connectionTimeout time.Duration,
	maxRetries int,
	retryDelay time.Duration,
) (*ConnectionManager, error) {

	return &ConnectionManager{
		certPath:          certPath,
		connectionTimeout: connectionTimeout,
		maxRetries:        maxRetries,
		retryDelay:        retryDelay,
	}, nil
}

func loadTLSConfig(certPath string) (*tls.Config, error) {
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

func (cm *ConnectionManager) retry(name string, fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= cm.maxRetries; attempt++ {
		err := fn()
		if err == nil {
			if attempt > 1 {
				log.Printf(
					"%s: подключение успешно после попытки %d/%d",
					name,
					attempt,
					cm.maxRetries,
				)
			}
			return nil
		}

		lastErr = err

		if attempt < cm.maxRetries {
			waitTime := cm.retryDelay * time.Duration(1<<uint(attempt-1))
			log.Printf(
				"%s поптыка %d/%d не удалась: %v, повтор через %v",
				name,
				attempt,
				cm.maxRetries,
				err,
				waitTime,
			)

			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("не удалось подключиться после %d попыток: %w", cm.maxRetries, lastErr)
}
