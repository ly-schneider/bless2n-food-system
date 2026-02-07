package app

import (
	"context"

	"backend/internal/blobstore"
	"backend/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewBlobStore(lc fx.Lifecycle, cfg config.Config, logger *zap.Logger) *blobstore.Client {
	if cfg.BlobStorage.AccountName == "" || cfg.BlobStorage.AccountKey == "" {
		logger.Warn("blob storage credentials not configured, image uploads disabled")
		return nil
	}

	client, err := blobstore.NewClient(
		cfg.BlobStorage.AccountName,
		cfg.BlobStorage.AccountKey,
		cfg.BlobStorage.Container,
		cfg.BlobStorage.BlobEndpoint,
	)
	if err != nil {
		logger.Error("failed to create blob storage client", zap.Error(err))
		return nil
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.EnsureContainer(ctx); err != nil {
				logger.Warn("failed to ensure blob container", zap.Error(err))
			}
			return nil
		},
	})

	return client
}
