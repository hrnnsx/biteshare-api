package handler

import (
	"errors"
	"net/http"
	"strconv"

	"biteshare-api/donation-api/dto"
	"biteshare-api/donation-api/repository"
	"biteshare-api/donation-api/service"

	"github.com/gin-gonic/gin"
)

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

// Create handles POST /donations
func (h *DonationHandler) Create(c *gin.Context) {
	var input dto.CreateDonationRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	result, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Donation request submitted successfully",
		"data":    result,
	})
}

// GetByID handles GET /donations/:id
func (h *DonationHandler) GetByID(c *gin.Context) {
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || requestID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid donation request ID",
		})
		return
	}

	result, err := h.service.GetByID(
		c.Request.Context(),
		int32(requestID),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// GetAll handles GET /donations
func (h *DonationHandler) GetAll(c *gin.Context) {
	status := c.Query("status")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid page parameter",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid limit parameter",
		})
		return
	}

	result, total, err := h.service.GetAll(
		c.Request.Context(),
		status,
		page,
		limit,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// Review handles PATCH /donations/:id/review
func (h *DonationHandler) Review(c *gin.Context) {
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || requestID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid donation request ID",
		})
		return
	}

	var input dto.ReviewDonationRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	err = h.service.Review(
		c.Request.Context(),
		int32(requestID),
		input,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Donation request reviewed successfully",
	})
}

// handleError converts service/repository errors into HTTP responses.
func (h *DonationHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrDonationNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Donation request not found",
		})

	case errors.Is(err, service.ErrInvalidDonationData),
		errors.Is(err, service.ErrInvalidStatus):
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})

	case errors.Is(err, service.ErrInvalidTransition):
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": "Donation request cannot be reviewed in its current status",
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Internal server error",
		})
	}
}
