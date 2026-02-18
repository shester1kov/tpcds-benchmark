package storage

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"tpcds_benchmark/pkg/config"
	"tpcds_benchmark/pkg/utils"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	client *minio.Client
	bucket string
	prefix string
}

func NewS3Storage(cfg *config.S3Config, certPath string) (*S3Storage, error) {
	tlsConfig, err := utils.LoadTLSConfig(certPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании tlsConfig: %v", err)
	}

	client, err := minio.New(
		cfg.Endpoint,
		&minio.Options{
			Creds: credentials.NewStaticV4(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
			Secure: cfg.UseSSL,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Region: cfg.Region,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("ошибка при создании конфигурации: %w", err)
	}

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
		prefix: cfg.Prefix,
	}, nil
}

func (s *S3Storage) Upload(filePath string) error {
	ctx := context.Background()

	info, err := s.client.FPutObject(
		ctx,
		s.bucket,
		fmt.Sprintf("%s/%s", s.prefix, filepath.Base(filePath)),
		filePath, minio.PutObjectOptions{
			ContentType: "text/csv",
		},
	)

	if err != nil {
		return fmt.Errorf("ошибка загрузки файла в s3 хранилище: %w", err)
	}

	log.Printf("файл успешно загружен: %d байт", info.Size)

	return nil
}
