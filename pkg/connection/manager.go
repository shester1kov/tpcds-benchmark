package connection

import (
	"fmt"
	"log"
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
