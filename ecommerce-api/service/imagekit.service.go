package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

// ImageKitUploader uploads product images and returns their hosted URL and file ID.
type ImageKitUploader struct {
	privateKey string
	uploadURL  string
	client     *http.Client
}

type imageKitResponse struct {
	URL    string `json:"url"`
	FileID string `json:"fileId"`
}

func NewImageKitUploader() *ImageKitUploader {
	privateKey := os.Getenv("IMAGEKIT_PRIVATE")
	if privateKey == "" {
		return nil
	}

	uploadURL := os.Getenv("IMAGEKIT_UPLOAD_URL")
	if uploadURL == "" {
		uploadURL = "https://upload.imagekit.io/api/v1/files/upload"
	}

	return &ImageKitUploader{
		privateKey: privateKey,
		uploadURL:  uploadURL,
		client:     http.DefaultClient,
	}
}

func (u *ImageKitUploader) Upload(ctx context.Context, file *multipart.FileHeader) (string, string, error) {
	input, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("open image: %w", err)
	}
	defer input.Close()

	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	errCh := make(chan error, 1)
	go func() {
		defer writer.Close()
		defer multipartWriter.Close()

		part, err := multipartWriter.CreateFormFile("file", file.Filename)
		if err != nil {
			errCh <- err
			return
		}
		if _, err = io.Copy(part, input); err != nil {
			errCh <- err
			return
		}
		if err = multipartWriter.WriteField("fileName", file.Filename); err != nil {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, u.uploadURL, reader)
	if err != nil {
		return "", "", fmt.Errorf("create ImageKit request: %w", err)
	}
	request.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	request.SetBasicAuth(u.privateKey, "")

	response, err := u.client.Do(request)
	if err != nil {
		return "", "", fmt.Errorf("upload image to ImageKit: %w", err)
	}
	defer response.Body.Close()

	if err := <-errCh; err != nil {
		return "", "", fmt.Errorf("read image: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", "", fmt.Errorf("ImageKit returned status %s", response.Status)
	}

	var result imageKitResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decode ImageKit response: %w", err)
	}
	if result.URL == "" || result.FileID == "" {
		return "", "", fmt.Errorf("ImageKit response is missing url or fileId")
	}

	return result.URL, result.FileID, nil
}

// Delete removes a file from ImageKit using its file ID.
func (u *ImageKitUploader) Delete(ctx context.Context, fileID string) error {
	requestURL := "https://api.imagekit.io/v1/files/" + url.PathEscape(fileID)
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURL, nil)
	if err != nil {
		return fmt.Errorf("create ImageKit delete request: %w", err)
	}
	request.SetBasicAuth(u.privateKey, "")

	response, err := u.client.Do(request)
	if err != nil {
		return fmt.Errorf("delete image from ImageKit: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return fmt.Errorf("ImageKit delete failed with status %s: %s", response.Status, string(body))
	}

	return nil
}

