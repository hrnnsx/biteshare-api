package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"biteshare-api/donation-api/dto"
	"biteshare-api/donation-api/repository"
	"biteshare-api/donation-api/service"

	"github.com/gin-gonic/gin"
)

const maxDonationRequestSize = 50 << 20 // 50 MB

type DonationHandler struct {
	service service.DonationService
}

func NewDonationHandler(
	donationService service.DonationService,
) *DonationHandler {
	return &DonationHandler{
		service: donationService,
	}
}

// Create handles a new food donation submission.
//
// @Summary Submit a food donation
// @Description Submit a donation request containing donor information and one or more food items. Each food item must include at least one image. Images are uploaded to ImageKit.
// @Tags Donations
// @Accept multipart/form-data
// @Produce json
// @Param payload formData string true "Donation information in JSON format"
// @Param foods[0].images formData file true "Food item image. Use the corresponding index for each food item; at least one image is required per item."
// @Success 201 {object} map[string]interface{} "Donation submitted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid payload or missing food image"
// @Failure 413 {object} map[string]interface{} "Request or image exceeds size limit"
// @Failure 415 {object} map[string]interface{} "Content type must be multipart/form-data"
// @Failure 502 {object} map[string]interface{} "Image upload failed"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/donations [post]
func (h *DonationHandler) Create(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")

	if !strings.HasPrefix(contentType, "multipart/form-data") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error": "content type must be multipart/form-data",
		})
		return
	}

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxDonationRequestSize,
	)

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		var maxBytesErr *http.MaxBytesError

		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "donation request exceeds maximum size of 50 MB",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid multipart request",
			"details": err.Error(),
		})
		return
	}

	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}

	// Read JSON payload.
	payload := c.PostForm("payload")

	if payload == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "payload field is required",
		})
		return
	}

	var input dto.CreateDonationRequest

	if err := json.Unmarshal([]byte(payload), &input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid donation payload",
			"details": err.Error(),
		})
		return
	}

	if len(input.Foods) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "at least one food item is required",
		})
		return
	}

	form := c.Request.MultipartForm

	foodImages := make([][]*multipart.FileHeader, len(input.Foods))

	for i := range input.Foods {
		fieldName := fmt.Sprintf("foods[%d].images", i)

		foodImages[i] = form.File[fieldName]

		if len(foodImages[i]) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf(
					"at least one image is required for food item %d",
					i+1,
				),
			})
			return
		}
	}

	donation, err := h.service.Create(
		c.Request.Context(),
		input,
		foodImages,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "donation submitted successfully",
		"data":    donation,
	})
}

// GetByID retrieves a donation request by its ID.
//
// @Summary Get donation details
// @Description Retrieve complete information about a donation request, including donor, donated foods, images, and pickup schedules.
// @Tags Donations
// @Produce json
// @Security BearerAuth
// @Param id path int true "Donation request ID"
// @Success 200 {object} map[string]interface{} "Donation details"
// @Failure 400 {object} map[string]interface{} "Invalid donation ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Donation request not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/donations/{id} [get]
func (h *DonationHandler) GetByID(c *gin.Context) {
	requestID, err := parseRequestID(c)
	if err != nil {
		h.handleError(c, service.ErrInvalidDonationData)
		return
	}

	donation, err := h.service.GetByID(c.Request.Context(), requestID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": donation})
}

// GetAll retrieves a paginated list of donation requests.
//
// @Summary Get all donations
// @Description Retrieve donation requests with optional status filtering and pagination.
// @Tags Donations
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by donation status" Enums(pending,approved,rejected,pickup_scheduled,picked_up,received)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of records per page" default(10)
// @Success 200 {object} map[string]interface{} "Donation list and pagination information"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/donations [get]
func (h *DonationHandler) GetAll(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		h.handleError(c, service.ErrInvalidDonationData)
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		h.handleError(c, service.ErrInvalidDonationData)
		return
	}

	donations, total, err := h.service.GetAll(
		c.Request.Context(),
		c.Query("status"),
		page,
		limit,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": donations,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// Review approves or rejects a donation request.
//
// @Summary Review a donation request
// @Description Admin reviews a pending donation request and changes its status to approved or rejected.
// @Tags Donations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Donation request ID"
// @Param request body dto.ReviewDonationRequest true "Donation review information"
// @Success 200 {object} map[string]interface{} "Donation reviewed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or status"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Donation request not found"
// @Failure 409 {object} map[string]interface{} "Invalid donation status transition"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/donations/{id}/review [patch]
func (h *DonationHandler) Review(c *gin.Context) {
	requestID, err := parseRequestID(c)
	if err != nil {
		h.handleError(c, service.ErrInvalidDonationData)
		return
	}

	var input dto.ReviewDonationRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	donation, err := h.service.Review(
		c.Request.Context(),
		requestID,
		input,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "donation reviewed successfully",
		"data":    donation,
	})
}

func (h *DonationHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrDonationNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "donation request not found",
		})

	case errors.Is(err, service.ErrInvalidDonationData):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid donation data",
		})

	case errors.Is(err, service.ErrInvalidStatus):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid donation status",
		})

	case errors.Is(err, service.ErrInvalidTransition):
		c.JSON(http.StatusConflict, gin.H{
			"error": "invalid donation status transition",
		})

	case errors.Is(err, service.ErrMissingFoodImage):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "every food item must have at least one image",
		})

	case errors.Is(err, service.ErrInvalidImage):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid image file",
		})

	case errors.Is(err, service.ErrInvalidImageType):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "only JPEG, PNG, and WebP images are allowed",
		})

	case errors.Is(err, service.ErrImageTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "each image must not exceed 5 MB",
		})

	case errors.Is(err, service.ErrImageUploadFailed):
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "failed to upload image to storage provider",
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}

func parseRequestID(c *gin.Context) (int32, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid request ID")
	}

	return int32(id), nil
}
