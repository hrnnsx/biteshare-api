package model

import "time"

type DonationRequest struct {
	ID          uint64         `json:"request_id"`
	DonorID     uint64         `json:"donor_id"`
	Status      DonationStatus `json:"status"`
	Notes       string         `json:"notes,omitempty"`
	SubmittedAt time.Time      `json:"submitted_at"`
}
