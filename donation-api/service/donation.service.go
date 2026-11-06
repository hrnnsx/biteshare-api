package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"biteshare-api/donation-api/dto"
	"biteshare-api/donation-api/model"
	"biteshare-api/donation-api/repository"
	"biteshare-api/internal/imagekit"
)

var (
	ErrInvalidDonationData = errors.New("invalid donation data")
	ErrInvalidStatus       = errors.New("invalid donation status")
	ErrInvalidTransition   = errors.New("invalid donation status transition")

	ErrMissingFoodImage  = errors.New("each donated food must have at least one image")
	ErrInvalidImage      = errors.New("invalid image file")
	ErrInvalidImageType  = errors.New("unsupported image type")
	ErrImageTooLarge     = errors.New("image exceeds maximum size of 5 MB")
	ErrImageUploadFailed = errors.New("image upload failed")
)

const (
	maxImageSize = 5 << 20 // 5 MB
	imageFolder  = "BiteShare/donations/food-images"
)

type ImageStorage interface {
	Upload(
		ctx context.Context,
		file io.Reader,
		fileName string,
		contentType string,
		folder string,
	) (*imagekit.UploadedFile, error)

	Delete(ctx context.Context, fileID string) error
}

type DonationService interface {
	Create(
		ctx context.Context,
		input dto.CreateDonationRequest,
		foodImages [][]*multipart.FileHeader,
	) (*model.DonationRequest, error)

	GetByID(ctx context.Context, requestID int32) (*model.DonationRequest, error)

	GetAll(
		ctx context.Context,
		status string,
		page, limit int,
	) ([]model.DonationRequest, int64, error)

	Review(
		ctx context.Context,
		requestID int32,
		input dto.ReviewDonationRequest,
	) (*model.DonationRequest, error)
}

type donationService struct {
	repo       repository.DonationRepository
	imageStore ImageStorage
}

func NewDonationService(
	repo repository.DonationRepository,
	imageStore ImageStorage,
) DonationService {
	return &donationService{
		repo:       repo,
		imageStore: imageStore,
	}
}

func (s *donationService) Create(
	ctx context.Context,
	input dto.CreateDonationRequest,
	foodImages [][]*multipart.FileHeader,
) (*model.DonationRequest, error) {
	if strings.TrimSpace(input.Donor.Name) == "" ||
		strings.TrimSpace(input.Donor.Email) == "" ||
		strings.TrimSpace(input.Donor.Phone) == "" ||
		strings.TrimSpace(input.Donor.Address) == "" ||
		len(input.Foods) == 0 {
		return nil, ErrInvalidDonationData
	}

	if s.imageStore == nil {
		return nil, ErrImageUploadFailed
	}

	if len(foodImages) != len(input.Foods) {
		return nil, ErrMissingFoodImage
	}

	donor := &model.Donor{
		Name:    strings.TrimSpace(input.Donor.Name),
		Email:   strings.TrimSpace(input.Donor.Email),
		Phone:   strings.TrimSpace(input.Donor.Phone),
		Address: strings.TrimSpace(input.Donor.Address),
	}

	request := &model.DonationRequest{
		Status:      model.DonationPending,
		Notes:       stringToPointer(input.Notes),
		SubmittedAt: time.Now(),
	}

	foods := make([]model.DonatedFood, 0, len(input.Foods))

	for i, foodInput := range input.Foods {
		if strings.TrimSpace(foodInput.FoodName) == "" ||
			foodInput.Quantity <= 0 ||
			strings.TrimSpace(foodInput.Unit) == "" {
			return nil, ErrInvalidDonationData
		}

		if len(foodImages[i]) == 0 {
			return nil, ErrMissingFoodImage
		}

		var expiryDate *time.Time

		if foodInput.ExpiryDate != nil &&
			strings.TrimSpace(*foodInput.ExpiryDate) != "" {
			parsedDate, err := time.Parse(
				"2006-01-02",
				strings.TrimSpace(*foodInput.ExpiryDate),
			)
			if err != nil {
				return nil, ErrInvalidDonationData
			}

			expiryDate = &parsedDate
		}

		foods = append(foods, model.DonatedFood{
			FoodName:    strings.TrimSpace(foodInput.FoodName),
			Quantity:    foodInput.Quantity,
			Unit:        strings.TrimSpace(foodInput.Unit),
			ExpiryDate:  expiryDate,
			Description: stringToPointer(&foodInput.Description),
		})
	}

	// Upload all food images before creating database records.
	var uploadedFileIDs []string

	cleanup := func() {
		s.cleanupUploadedImages(ctx, uploadedFileIDs)
	}

	for foodIndex, files := range foodImages {
		for imageIndex, fileHeader := range files {
			contentType, err := validateImageFile(fileHeader)
			if err != nil {
				cleanup()
				return nil, err
			}

			file, err := fileHeader.Open()
			if err != nil {
				cleanup()
				return nil, fmt.Errorf("%w: %v", ErrInvalidImage, err)
			}

			fileName := imageFileName(foodIndex, imageIndex, contentType)

			uploaded, uploadErr := s.imageStore.Upload(
				ctx,
				file,
				fileName,
				contentType,
				imageFolder,
			)

			closeErr := file.Close()

			if uploadErr != nil {
				cleanup()
				return nil, fmt.Errorf("%w: %v", ErrImageUploadFailed, uploadErr)
			}

			if closeErr != nil {
				if uploaded != nil && uploaded.FileID != "" {
					uploadedFileIDs = append(uploadedFileIDs, uploaded.FileID)
				}
				cleanup()
				return nil, fmt.Errorf("close uploaded image: %w", closeErr)
			}

			if uploaded == nil || uploaded.FileID == "" || uploaded.URL == "" {
				cleanup()
				return nil, ErrImageUploadFailed
			}

			uploadedFileIDs = append(uploadedFileIDs, uploaded.FileID)

			fileID := uploaded.FileID

			foods[foodIndex].Images = append(
				foods[foodIndex].Images,
				model.DonationImage{
					ImageURL:  uploaded.URL,
					FileID:    &fileID,
					IsPrimary: imageIndex == 0,
					CreatedAt: time.Now(),
				},
			)
		}
	}

	// Save donor, request, foods, and image metadata in one DB transaction.
	if err := s.repo.Create(ctx, donor, request, foods); err != nil {
		cleanup()
		return nil, err
	}

	// Do not delete uploaded files if this read fails:
	// the database transaction has already succeeded.
	return s.repo.FindByID(ctx, request.ID)
}

func validateImageFile(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil || fileHeader.Size <= 0 {
		return "", ErrInvalidImage
	}

	if fileHeader.Size > maxImageSize {
		return "", ErrImageTooLarge
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidImage, err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)

	if err != nil && err != io.EOF {
		return "", fmt.Errorf("%w: %v", ErrInvalidImage, err)
	}

	contentType := http.DetectContentType(buffer[:n])

	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
		return contentType, nil
	default:
		return "", ErrInvalidImageType
	}
}

func imageFileName(foodIndex, imageIndex int, contentType string) string {
	extension := ".jpg"

	switch contentType {
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	}

	return fmt.Sprintf(
		"food-%d-image-%d%s",
		foodIndex+1,
		imageIndex+1,
		extension,
	)
}

func (s *donationService) cleanupUploadedImages(
	ctx context.Context,
	fileIDs []string,
) {
	if len(fileIDs) == 0 {
		return
	}

	// Cleanup should still run even if the original request was cancelled.
	cleanupCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		15*time.Second,
	)
	defer cancel()

	for _, fileID := range fileIDs {
		if err := s.imageStore.Delete(cleanupCtx, fileID); err != nil {
			log.Printf(
				"failed to cleanup ImageKit file %s: %v",
				fileID,
				err,
			)
		}
	}
}

func (s *donationService) GetByID(
	ctx context.Context,
	requestID int32,
) (*model.DonationRequest, error) {
	if requestID <= 0 {
		return nil, ErrInvalidDonationData
	}

	return s.repo.FindByID(ctx, requestID)
}

func (s *donationService) GetAll(
	ctx context.Context,
	status string,
	page, limit int,
) ([]model.DonationRequest, int64, error) {
	if status != "" {
		validStatuses := map[string]bool{
			string(model.DonationPending):         true,
			string(model.DonationApproved):        true,
			string(model.DonationRejected):        true,
			string(model.DonationPickupScheduled): true,
			string(model.DonationPickedUp):        true,
			string(model.DonationReceived):        true,
		}

		if !validStatuses[status] {
			return nil, 0, ErrInvalidStatus
		}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.FindAll(ctx, status, page, limit)
}

func (s *donationService) Review(
	ctx context.Context,
	requestID int32,
	input dto.ReviewDonationRequest,
) (*model.DonationRequest, error) {
	if requestID <= 0 {
		return nil, ErrInvalidDonationData
	}

	var newStatus model.DonationStatus

	switch input.Status {
	case dto.ReviewApproved:
		newStatus = model.DonationApproved
	case dto.ReviewRejected:
		newStatus = model.DonationRejected
	default:
		return nil, ErrInvalidStatus
	}

	donation, err := s.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if donation.Status != model.DonationPending {
		return nil, ErrInvalidTransition
	}

	if err := s.repo.UpdateStatus(ctx, requestID, newStatus); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, requestID)
}

func stringToPointer(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

var _ DonationService = (*donationService)(nil)
