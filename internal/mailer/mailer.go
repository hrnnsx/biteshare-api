package mailer

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/resend/resend-go/v4"
)

type Mailer interface {
	SendEmail(to []string, subject, htmlBody string) error
	SendDonationStatusEmail(toEmail, donorName, status, notes string) error
	SendOrderReceiptEmail(toEmail, buyerName, orderNumber string, totalAmount float64) error
}

type resendMailer struct {
	client    *resend.Client
	fromEmail string
}

func New() Mailer {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Println("[Mailer] Warning: RESEND_API_KEY is not configured")
	}

	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev"
	}

	return &resendMailer{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
	}
}

func (m *resendMailer) SendEmail(to []string, subject, htmlBody string) error {
	if m.client == nil {
		return errors.New("resend client is not initialized")
	}

	params := &resend.SendEmailRequest{
		From:    m.fromEmail,
		To:      to,
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := m.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend send error: %w", err)
	}

	log.Printf("[Mailer] Email sent successfully. ID: %s", sent.Id)
	return nil
}

// Helper untuk Person 1 (Donation API)
func (m *resendMailer) SendDonationStatusEmail(toEmail, donorName, status, notes string) error {
	subject := fmt.Sprintf("BiteShare - Status Donasi Makanan Anda: %s", status)
	htmlBody := fmt.Sprintf(`
		<h2>Halo %s,</h2>
		<p>Status pengajuan donasi makanan Anda saat ini adalah: <strong>%s</strong>.</p>
		<p>Catatan Tim: %s</p>
		<br>
		<p>Terima kasih atas kontribusi Anda dalam mengurangi pemborosan makanan!</p>
		<p><strong>BiteShare Team</strong></p>
	`, donorName, status, notes)

	return m.SendEmail([]string{toEmail}, subject, htmlBody)
}

// Helper untuk Person 2 (E-Commerce API)
func (m *resendMailer) SendOrderReceiptEmail(toEmail, buyerName, orderNumber string, totalAmount float64) error {
	subject := fmt.Sprintf("BiteShare - Bukti Pesanan #%s", orderNumber)
	htmlBody := fmt.Sprintf(`
		<h2>Halo %s,</h2>
		<p>Terima kasih telah berbelanja di BiteShare.</p>
		<p>Nomor Pesanan: <strong>%s</strong></p>
		<p>Total Pembayaran: <strong>Rp %.2f</strong></p>
		<p>Pesanan Anda sedang dipersiapkan untuk pengiriman.</p>
		<br>
		<p><strong>BiteShare Team</strong></p>
	`, buyerName, orderNumber, totalAmount)

	return m.SendEmail([]string{toEmail}, subject, htmlBody)
}
