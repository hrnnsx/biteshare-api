package handler

import (
	"errors"
	"net/http"

	"biteshare-api/donation-api/dto"
	"biteshare-api/donation-api/repository"
	"biteshare-api/donation-api/service"
	"biteshare-api/internal/auth"

	"github.com/gin-gonic/gin"
)

type PickupScheduleHandler struct {
	service service.PickupScheduleService
}

func NewPickupScheduleHandler(
	pickupScheduleService service.PickupScheduleService,
) *PickupScheduleHandler {
	return &PickupScheduleHandler{service: pickupScheduleService}
}

func (h *PickupScheduleHandler) SchedulePickup(c *gin.Context) {
	requestID, err := parseRequestID(c)
	if err != nil {
		h.handleError(c, service.ErrInvalidScheduleData)
		return
	}

	adminID, _ := auth.AdminID(c)

	var input dto.CreatePickupScheduleRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	schedule, err := h.service.SchedulePickup(
		c.Request.Context(),
		requestID,
		adminID,
		input,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "pickup scheduled successfully",
		"data":    schedule,
	})
}

func (h *PickupScheduleHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrDonationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "donation request not found"})

	case errors.Is(err, repository.ErrDonationNotApproved):
		c.JSON(http.StatusConflict, gin.H{
			"error": "donation request must be approved before scheduling pickup",
		})

	case errors.Is(err, service.ErrInvalidScheduleData):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pickup schedule data"})

	case errors.Is(err, service.ErrInvalidAdminID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin ID"})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
