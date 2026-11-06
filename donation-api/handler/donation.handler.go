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
	return &DonationHandler{service: donationService}
}

func (h *DonationHandler) Create(c *gin.Context) {
	var input dto.CreateDonationRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	donation, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "donation submitted successfully",
		"data":    donation,
	})
}

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

	donation, err := h.service.Review(c.Request.Context(), requestID, input)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "donation request not found"})

	case errors.Is(err, service.ErrInvalidDonationData):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid donation data"})

	case errors.Is(err, service.ErrInvalidStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid donation status"})

	case errors.Is(err, service.ErrInvalidTransition):
		c.JSON(http.StatusConflict, gin.H{"error": "invalid donation status transition"})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func parseRequestID(c *gin.Context) (int32, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid request ID")
	}

	return int32(id), nil
}
