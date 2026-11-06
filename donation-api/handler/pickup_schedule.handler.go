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

// SchedulePickup creates a pickup schedule for an approved donation.
//
// @Summary Schedule a donation pickup
// @Description Admin schedules the pickup date and time for an approved donation request. The donor receives an email notification after the schedule is successfully created.
// @Tags Pickup Schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Donation request ID"
// @Param request body dto.CreatePickupScheduleRequest true "Pickup schedule information"
// @Success 201 {object} map[string]interface{} "Pickup scheduled successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or schedule data"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Donation request not found"
// @Failure 409 {object} map[string]interface{} "Donation must be approved before scheduling pickup"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/donations/{id}/schedules [post]
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
