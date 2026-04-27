package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/develoFavour/e-voting-system-be/internal/config"
)

const brevoSMTPURL = "https://api.brevo.com/v3/smtp/email"

type brevoEmailRequest struct {
	Sender      brevoSender      `json:"sender"`
	To          []brevoRecipient `json:"to"`
	Subject     string           `json:"subject"`
	HTMLContent string           `json:"htmlContent"`
}

type brevoSender struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type brevoRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type BrevoEmailService struct {
	apiKey      string
	senderEmail string
	senderName  string
	client      *http.Client
	frontendURL string
}

func NewBrevoEmailService(cfg *config.Config) *BrevoEmailService {
	return &BrevoEmailService{
		apiKey:      strings.TrimSpace(cfg.BrevoAPIKey),
		senderEmail: strings.TrimSpace(cfg.SenderEmail),
		senderName:  strings.TrimSpace(cfg.SenderName),
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		frontendURL: strings.TrimSuffix(strings.TrimSpace(cfg.FrontendURL), "/"),
	}
}

func (s *BrevoEmailService) IsConfigured() bool {
	return s.apiKey != "" && s.senderEmail != ""
}

func (s *BrevoEmailService) SendPasswordResetEmail(userEmail, userName, resetToken string) error {
	if !s.IsConfigured() {
		return fmt.Errorf("brevo email service is not configured")
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, resetToken)
	safeName := html.EscapeString(strings.TrimSpace(userName))

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Password Reset</title>
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #1f2937; background: #f3f4f6; }
    .container { max-width: 600px; margin: 0 auto; padding: 24px; }
    .card { background: #ffffff; border-radius: 12px; overflow: hidden; }
    .header { background: #0ea5e9; color: white; padding: 24px; text-align: center; }
    .content { padding: 32px 24px; }
    .button { display: inline-block; background: #0ea5e9; color: white !important; text-decoration: none; padding: 12px 24px; border-radius: 8px; font-weight: 600; }
    .footer { color: #6b7280; font-size: 12px; margin-top: 24px; text-align: center; }
    .note { background: #eff6ff; border: 1px solid #bae6fd; padding: 16px; border-radius: 8px; margin-top: 24px; }
  </style>
</head>
<body>
  <div class="container">
    <div class="card">
      <div class="header">
        <h1>Password Reset Request</h1>
      </div>
      <div class="content">
        <p>Hello %s,</p>
        <p>We received a request to reset your password for the Hallmark e-voting system.</p>
        <p style="text-align:center; margin: 32px 0;">
          <a href="%s" class="button">Reset Password</a>
        </p>
        <div class="note">
          This link expires in 1 hour. If you did not request a password reset, you can safely ignore this email.
        </div>
      </div>
    </div>
    <div class="footer">
      This is an automated message. Please do not reply.
    </div>
  </div>
</body>
</html>`, safeName, resetURL)

	return s.sendEmail(userEmail, "Reset Your Hallmark E-Voting Password", htmlContent)
}

func (s *BrevoEmailService) SendAccreditationRejectedEmail(userEmail, userName, rejectionReason string) error {
	if !s.IsConfigured() {
		return fmt.Errorf("brevo email service is not configured")
	}

	safeName := html.EscapeString(strings.TrimSpace(userName))
	safeReason := html.EscapeString(strings.TrimSpace(rejectionReason))

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Accreditation Update</title>
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #1f2937; background: #f3f4f6; }
    .container { max-width: 600px; margin: 0 auto; padding: 24px; }
    .card { background: #ffffff; border-radius: 12px; overflow: hidden; }
    .header { background: #ef4444; color: white; padding: 24px; text-align: center; }
    .content { padding: 32px 24px; }
    .reason { background: #fef2f2; border: 1px solid #fecaca; padding: 16px; border-radius: 8px; margin: 24px 0; }
    .footer { color: #6b7280; font-size: 12px; margin-top: 24px; text-align: center; }
  </style>
</head>
<body>
  <div class="container">
    <div class="card">
      <div class="header">
        <h1>Accreditation Request Update</h1>
      </div>
      <div class="content">
        <p>Hello %s,</p>
        <p>We reviewed your voter accreditation request and we are unable to approve it at this time.</p>
        <div class="reason">
          <strong>Reason provided by the administrator:</strong>
          <p>%s</p>
        </div>
        <p>Please review the reason above and submit a new accreditation request when you are ready.</p>
      </div>
    </div>
    <div class="footer">
      This is an automated message. Please do not reply.
    </div>
  </div>
</body>
</html>`, safeName, safeReason)

	return s.sendEmail(userEmail, "Your Hallmark E-Voting Accreditation Request Was Rejected", htmlContent)
}

func (s *BrevoEmailService) SendAccreditationApprovedEmail(userEmail, userName string) error {
	if !s.IsConfigured() {
		return fmt.Errorf("brevo email service is not configured")
	}

	safeName := html.EscapeString(strings.TrimSpace(userName))
	loginURL := fmt.Sprintf("%s/login", s.frontendURL)

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Accreditation Approved</title>
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #1f2937; background: #f3f4f6; }
    .container { max-width: 600px; margin: 0 auto; padding: 24px; }
    .card { background: #ffffff; border-radius: 12px; overflow: hidden; }
    .header { background: #10b981; color: white; padding: 24px; text-align: center; }
    .content { padding: 32px 24px; }
    .button { display: inline-block; background: #10b981; color: white !important; text-decoration: none; padding: 12px 24px; border-radius: 8px; font-weight: 600; }
    .footer { color: #6b7280; font-size: 12px; margin-top: 24px; text-align: center; }
  </style>
</head>
<body>
  <div class="container">
    <div class="card">
      <div class="header">
        <h1>Accreditation Approved</h1>
      </div>
      <div class="content">
        <p>Hello %s,</p>
        <p>Your voter accreditation request has been approved successfully.</p>
        <p>You can now sign in to the Hallmark e-voting system with your matric number and password.</p>
        <p style="text-align:center; margin: 32px 0;">
          <a href="%s" class="button">Go to Login</a>
        </p>
        <p>We wish you a smooth voting experience.</p>
      </div>
    </div>
    <div class="footer">
      This is an automated message. Please do not reply.
    </div>
  </div>
</body>
</html>`, safeName, loginURL)

	return s.sendEmail(userEmail, "Your Hallmark E-Voting Accreditation Has Been Approved", htmlContent)
}

func (s *BrevoEmailService) sendEmail(toEmail, subject, htmlContent string) error {
	payload := brevoEmailRequest{
		Sender: brevoSender{
			Name:  s.senderName,
			Email: s.senderEmail,
		},
		To: []brevoRecipient{
			{Email: strings.TrimSpace(toEmail)},
		},
		Subject:     subject,
		HTMLContent: htmlContent,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, brevoSMTPURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("brevo API returned status %d", resp.StatusCode)
	}

	return nil
}
