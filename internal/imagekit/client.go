package imagekit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	ik "github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
)

var (
	ErrMissingPrivateKey = errors.New("imagekit private key is not configured")
	ErrInvalidUpload     = errors.New("invalid image upload data")
)

type Client struct {
	sdk ik.Client
}

type UploadedFile struct {
	FileID string
	URL    string
}

func NewClient() (*Client, error) {
	privateKey := strings.TrimSpace(os.Getenv("IMAGEKIT_PRIVATE"))

	if privateKey == "" {
		return nil, ErrMissingPrivateKey
	}

	sdk := ik.NewClient(
		option.WithPrivateKey(privateKey),
	)

	return &Client{
		sdk: sdk,
	}, nil
}

func (c *Client) Upload(
	ctx context.Context,
	file io.Reader,
	fileName string,
	contentType string,
	folder string,
) (*UploadedFile, error) {
	if file == nil || strings.TrimSpace(fileName) == "" {
		return nil, ErrInvalidUpload
	}

	fileName = strings.TrimSpace(fileName)

	uploadFile := ik.NewFile(
		file,
		fileName,
		contentType,
	)

	params := ik.FileUploadParams{
		File:              uploadFile,
		FileName:          fileName,
		UseUniqueFileName: ik.Bool(true),
	}

	if strings.TrimSpace(folder) != "" {
		params.Folder = ik.String(folder)
	}

	result, err := c.sdk.Files.Upload(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("upload image to ImageKit: %w", err)
	}

	if result.FileID == "" || result.URL == "" {
		return nil, errors.New("ImageKit returned an incomplete upload response")
	}

	return &UploadedFile{
		FileID: result.FileID,
		URL:    result.URL,
	}, nil
}

func (c *Client) Delete(
	ctx context.Context,
	fileID string,
) error {
	fileID = strings.TrimSpace(fileID)

	if fileID == "" {
		return errors.New("imagekit file ID is required")
	}

	if err := c.sdk.Files.Delete(ctx, fileID); err != nil {
		return fmt.Errorf("delete image from ImageKit: %w", err)
	}

	return nil
}
