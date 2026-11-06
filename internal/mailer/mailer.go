package mailer

import (
	"errors"
	"fmt"
	"html"
	"log"
	"os"
	"strings"
	"time"

	"github.com/resend/resend-go/v4"
)

type Mailer interface {
	SendEmail(to []string, subject, htmlBody string) error
	SendDonationStatusEmail(toEmail, donorName, status, notes string) error
	SendDonationSubmittedEmail(toEmail, donorName string, requestID int32) error
	SendPickupScheduledEmail(
		toEmail, donorName string,
		requestID int32,
		pickupDate time.Time,
		pickupTime string,
	) error
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

func (m *resendMailer) SendEmail(
	to []string,
	subject, htmlBody string,
) error {
	if m.client == nil {
		return errors.New("resend client is not initialized")
	}

	if len(to) == 0 {
		return errors.New("recipient email is required")
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

func (m *resendMailer) SendDonationSubmittedEmail(
	toEmail, donorName string,
	requestID int32,
) error {
	subject := fmt.Sprintf(
		"BiteShare - Konfirmasi Pengajuan Donasi #%d",
		requestID,
	)

	htmlBody := fmt.Sprintf(`
		<h2>Halo %s,</h2>
		<p>Terima kasih telah mengajukan donasi makanan melalui BiteShare.</p>
		<p>Pengajuan donasi Anda telah berhasil kami terima dan saat ini
		sedang menunggu proses peninjauan oleh tim BiteShare.</p>
		<p>Nomor pengajuan: <strong>#%d</strong></p>
		<p>Status: <strong>Menunggu peninjauan</strong></p>
		<br>
		<p>Terima kasih telah membantu mengurangi pemborosan makanan.</p>
		<p><strong>BiteShare Team</strong></p>
	`, html.EscapeString(donorName), requestID)

	return m.SendEmail([]string{toEmail}, subject, htmlBody)
}

func (m *resendMailer) SendDonationStatusEmail(
	toEmail, donorName, status, notes string,
) error {
	subject := fmt.Sprintf(
		"BiteShare - Status Donasi Makanan Anda: %s",
		status,
	)

	notesHTML := ""
	if strings.TrimSpace(notes) != "" {
		notesHTML = fmt.Sprintf(
			"<p>Catatan Tim: %s</p>",
			html.EscapeString(notes),
		)
	}

	htmlBody := fmt.Sprintf(`
		<h2>Halo %s,</h2>
		<p>Status pengajuan donasi makanan Anda saat ini adalah:
		<strong>%s</strong>.</p>
		%s
		<br>
		<p>Terima kasih atas kontribusi Anda dalam mengurangi
		pemborosan makanan!</p>
		<p><strong>BiteShare Team</strong></p>
	`,
		html.EscapeString(donorName),
		html.EscapeString(status),
		notesHTML,
	)

	return m.SendEmail([]string{toEmail}, subject, htmlBody)
}

func (m *resendMailer) SendPickupScheduledEmail(
	toEmail, donorName string,
	requestID int32,
	pickupDate time.Time,
	pickupTime string,
) error {
	subject := fmt.Sprintf(
		"BiteShare - Jadwal Pickup Donasi #%d",
		requestID,
	)

	formattedDate := pickupDate.Format("02 January 2006")

	htmlBody := fmt.Sprintf(`
		<h2>Halo %s,</h2>
		<p>Donasi makanan Anda telah dijadwalkan untuk dijemput.</p>
		<p>Berikut detail jadwal pickup:</p>
		<ul>
			<li>Nomor pengajuan: <strong>#%d</strong></li>
			<li>Tanggal: <strong>%s</strong></li>
			<li>Waktu: <strong>%s</strong></li>
		</ul>
		<p>Mohon pastikan makanan telah siap ketika tim BiteShare tiba.</p>
		<br>
		<p>Terima kasih telah menjadi bagian dari BiteShare.</p>
		<p><strong>BiteShare Team</strong></p>
	`,
		html.EscapeString(donorName),
		requestID,
		html.EscapeString(formattedDate),
		html.EscapeString(pickupTime),
	)

	return m.SendEmail([]string{toEmail}, subject, htmlBody)
}

func (m *resendMailer) SendOrderReceiptEmail(
	toEmail, buyerName, orderNumber string,
	totalAmount float64,
) error {
	subject := fmt.Sprintf("BiteShare - Bukti Pesanan #%s", orderNumber)

	htmlBody := fmt.Sprintf(`
		<h2>Halo %s,</h2>
		<p>Terima kasih telah berbelanja di BiteShare.</p>
		<p>Nomor Pesanan: <strong>%s</strong></p>
		<p>Total Pembayaran: <strong>Rp %.2f</strong></p>
		<p>Pesanan Anda sedang dipersiapkan untuk pengiriman.</p>
		<br>
		<p><strong>BiteShare Team</strong></p>
	`,
		html.EscapeString(buyerName),
		html.EscapeString(orderNumber),
		totalAmount,
	)

	return m.SendEmail([]string{toEmail}, subject, htmlBody)
}
