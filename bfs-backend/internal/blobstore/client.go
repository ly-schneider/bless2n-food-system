package blobstore

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/google/uuid"
)

type Client struct {
	client        *azblob.Client
	containerName string
	blobEndpoint  string
}

func NewClient(accountName, accountKey, containerName, blobEndpoint string) (*Client, error) {
	if blobEndpoint == "" {
		blobEndpoint = fmt.Sprintf("https://%s.blob.core.windows.net", accountName)
	}

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("blobstore: invalid credentials: %w", err)
	}

	client, err := azblob.NewClientWithSharedKeyCredential(blobEndpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("blobstore: failed to create client: %w", err)
	}

	return &Client{
		client:        client,
		containerName: containerName,
		blobEndpoint:  strings.TrimRight(blobEndpoint, "/"),
	}, nil
}

func (c *Client) Upload(ctx context.Context, data io.ReadSeeker, contentType string) (string, error) {
	blobName := uuid.New().String()

	cacheControl := "public, max-age=31536000, immutable"
	opts := &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType:  &contentType,
			BlobCacheControl: &cacheControl,
		},
	}

	_, err := c.client.UploadStream(ctx, c.containerName, blobName, data, opts)
	if err != nil {
		return "", fmt.Errorf("blobstore: upload failed: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", c.blobEndpoint, c.containerName, blobName)
	return url, nil
}

func (c *Client) Delete(ctx context.Context, blobURL string) error {
	prefix := fmt.Sprintf("%s/%s/", c.blobEndpoint, c.containerName)
	if !strings.HasPrefix(blobURL, prefix) {
		return nil
	}
	blobName := strings.TrimPrefix(blobURL, prefix)

	_, err := c.client.DeleteBlob(ctx, c.containerName, blobName, nil)
	if err != nil {
		return fmt.Errorf("blobstore: delete failed: %w", err)
	}
	return nil
}

func (c *Client) EnsureContainer(ctx context.Context) error {
	access := container.PublicAccessTypeBlob
	_, err := c.client.CreateContainer(ctx, c.containerName, &azblob.CreateContainerOptions{
		Access: &access,
	})
	if err != nil && !strings.Contains(err.Error(), "ContainerAlreadyExists") {
		return fmt.Errorf("blobstore: failed to create container: %w", err)
	}
	return nil
}
