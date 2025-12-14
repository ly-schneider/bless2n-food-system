package service

import (
	"backend/internal/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	plunkSendEndpoint = "https://api.useplunk.com/v1/send"
)

type EmailService interface {
	SendLoginEmail(ctx context.Context, to, code, ip, ua string, codeTTL time.Duration) error
	PreviewLoginEmail(ctx context.Context, to, code, ip, ua string, codeTTL time.Duration) (subject, text, html string, err error)
	SendEmailChangeVerification(ctx context.Context, toNewEmail, code, ip, ua string, codeTTL time.Duration) error
	PreviewEmailChangeVerification(ctx context.Context, toNewEmail, code, ip, ua string, codeTTL time.Duration) (subject, text, html string, err error)
	SendAdminInvite(ctx context.Context, to, token string, expiresAt time.Time) error
	PreviewAdminInvite(ctx context.Context, token string, ttl time.Duration) (subject, text, html string, err error)
}

type emailService struct {
	cfg      config.Config
	htmlTmpl *template.Template
	textTmpl *template.Template
	client   *http.Client
}

func NewEmailService(cfg config.Config) EmailService {
	return &emailService{
		cfg:      cfg,
		htmlTmpl: template.Must(template.New("login_html").Parse(loginEmailHTML)),
		textTmpl: template.Must(template.New("login_text").Parse(loginEmailText)),
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

type loginData struct {
	Brand       string
	Code        string
	CodeTTL     string
	IP          string
	UA          string
	SupportNote string
}

func (e *emailService) SendLoginEmail(ctx context.Context, to, code, ip, ua string, codeTTL time.Duration) error {
	data := loginData{
		Brand:       "BlessThun Food",
		Code:        code,
		CodeTTL:     friendlyTTL(codeTTL),
		IP:          ip,
		UA:          ua,
		SupportNote: "Wir werden dich niemals nach deinem Code fragen.",
	}

	var htmlBody strings.Builder
	var textBody strings.Builder
	if err := e.htmlTmpl.Execute(&htmlBody, data); err != nil {
		return err
	}
	if err := e.textTmpl.Execute(&textBody, data); err != nil {
		return err
	}

	subject := "Dein Anmeldecode"
	if err := e.send(ctx, to, subject, textBody.String(), htmlBody.String()); err != nil {
		return err
	}
	return nil
}

func (e *emailService) PreviewLoginEmail(ctx context.Context, to, code, ip, ua string, codeTTL time.Duration) (string, string, string, error) {
	data := loginData{
		Brand:       "BlessThun Food",
		Code:        code,
		CodeTTL:     friendlyTTL(codeTTL),
		IP:          ip,
		UA:          ua,
		SupportNote: "Wir werden dich niemals nach deinem Code fragen.",
	}
	var htmlBody strings.Builder
	var textBody strings.Builder
	if err := e.htmlTmpl.Execute(&htmlBody, data); err != nil {
		return "", "", "", err
	}
	if err := e.textTmpl.Execute(&textBody, data); err != nil {
		return "", "", "", err
	}
	return "Dein Anmeldecode", textBody.String(), htmlBody.String(), nil
}

type emailChangeData struct {
	Brand    string
	Code     string
	CodeTTL  string
	NewEmail string
	IP       string
	UA       string
}

func (e *emailService) SendEmailChangeVerification(ctx context.Context, toNewEmail, code, ip, ua string, codeTTL time.Duration) error {
	data := emailChangeData{
		Brand:    "BlessThun Food",
		Code:     code,
		CodeTTL:  friendlyTTL(codeTTL),
		NewEmail: toNewEmail,
		IP:       ip,
		UA:       ua,
	}
	htmlT := template.Must(template.New("email_change_html").Parse(emailChangeHTML))
	textT := template.Must(template.New("email_change_text").Parse(emailChangeText))
	var htmlBody, textBody strings.Builder
	if err := htmlT.Execute(&htmlBody, data); err != nil {
		return err
	}
	if err := textT.Execute(&textBody, data); err != nil {
		return err
	}
	subject := "E-Mail Änderung bestätigen"
	if err := e.send(ctx, toNewEmail, subject, textBody.String(), htmlBody.String()); err != nil {
		return err
	}
	return nil
}

func (e *emailService) PreviewEmailChangeVerification(ctx context.Context, toNewEmail, code, ip, ua string, codeTTL time.Duration) (string, string, string, error) {
	data := emailChangeData{
		Brand:    "BlessThun Food",
		Code:     code,
		CodeTTL:  friendlyTTL(codeTTL),
		NewEmail: toNewEmail,
		IP:       ip,
		UA:       ua,
	}
	htmlT := template.Must(template.New("email_change_html").Parse(emailChangeHTML))
	textT := template.Must(template.New("email_change_text").Parse(emailChangeText))
	var htmlBody, textBody strings.Builder
	if err := htmlT.Execute(&htmlBody, data); err != nil {
		return "", "", "", err
	}
	if err := textT.Execute(&textBody, data); err != nil {
		return "", "", "", err
	}
	return "E-Mail Änderung bestätigen", textBody.String(), htmlBody.String(), nil
}

type inviteData struct {
	Brand     string
	AcceptURL string
	TTL       string
}

func (e *emailService) SendAdminInvite(ctx context.Context, to, token string, expiresAt time.Time) error {
	// Build accept URL; public landing page handles acceptance
	base := strings.TrimRight(e.cfg.App.PublicBaseURL, "/")
	acceptURL := base + "/invite/accept?token=" + token
	ttl := time.Until(expiresAt)
	data := inviteData{Brand: "BlessThun Food", AcceptURL: acceptURL, TTL: friendlyTTL(ttl)}
	htmlT := template.Must(template.New("invite_html").Parse(adminInviteHTML))
	textT := template.Must(template.New("invite_text").Parse(adminInviteText))
	var htmlBody, textBody strings.Builder
	if err := htmlT.Execute(&htmlBody, data); err != nil {
		return err
	}
	if err := textT.Execute(&textBody, data); err != nil {
		return err
	}
	subject := "Einladung als Admin"
	if err := e.send(ctx, to, subject, textBody.String(), htmlBody.String()); err != nil {
		return err
	}
	return nil
}

func (e *emailService) PreviewAdminInvite(ctx context.Context, token string, ttl time.Duration) (string, string, string, error) {
	if token == "" {
		token = "preview-token"
	}
	base := strings.TrimRight(e.cfg.App.PublicBaseURL, "/")
	acceptURL := base + "/invite/accept?token=" + token
	data := inviteData{Brand: "BlessThun Food", AcceptURL: acceptURL, TTL: friendlyTTL(ttl)}
	htmlT := template.Must(template.New("invite_html").Parse(adminInviteHTML))
	textT := template.Must(template.New("invite_text").Parse(adminInviteText))
	var htmlBody, textBody strings.Builder
	if err := htmlT.Execute(&htmlBody, data); err != nil {
		return "", "", "", err
	}
	if err := textT.Execute(&textBody, data); err != nil {
		return "", "", "", err
	}
	return "Einladung als Admin", textBody.String(), htmlBody.String(), nil
}

type plunkSendRequest struct {
	To         string            `json:"to"`
	Subject    string            `json:"subject"`
	Body       string            `json:"body"`
	Subscribed bool              `json:"subscribed"`
	Name       string            `json:"name,omitempty"`
	From       string            `json:"from,omitempty"`
	Reply      string            `json:"reply,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
}

func (e *emailService) send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	fromName := e.cfg.Plunk.FromName
	fromEmail := e.cfg.Plunk.FromEmail
	replyTo := e.cfg.Plunk.ReplyTo

	body := strings.TrimSpace(htmlBody)
	if body == "" {
		body = textBody
	}

	payload := plunkSendRequest{
		To:         to,
		Subject:    subject,
		Body:       body,
		Subscribed: false,
		Name:       fromName,
		From:       fromEmail,
		Reply:      replyTo,
	}
	if textBody != "" {
		payload.Headers = map[string]string{"X-Text-Version": textBody}
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, plunkSendEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.cfg.Plunk.APIKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 10_240))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(respBody) > 0 {
			return fmt.Errorf("plunk send failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody)))
		}
		return fmt.Errorf("plunk send failed: status=%d", resp.StatusCode)
	}

	log.Printf("Sent email to %s via Plunk", to)
	return nil
}

func friendlyTTL(d time.Duration) string {
	// Make output human friendly for emails: include days, hours, minutes; drop seconds.
	if d < 0 {
		d = -d
	}
	// Avoid ugly fractional seconds like 59.997s
	d = d.Truncate(time.Minute)

	days := int(d / (24 * time.Hour))
	d -= time.Duration(days) * 24 * time.Hour
	hours := int(d / time.Hour)
	d -= time.Duration(hours) * time.Hour
	minutes := int(d / time.Minute)

	parts := make([]string, 0, 3)
	if days > 0 {
		if days == 1 {
			parts = append(parts, "1 Tag")
		} else {
			// Dative plural because templates use "in {{.TTL}}"
			parts = append(parts, fmt.Sprintf("%d Tagen", days))
		}
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d Std", hours))
	}
	if minutes > 0 || len(parts) == 0 {
		// Show minutes even if zero when there are no larger units
		parts = append(parts, fmt.Sprintf("%d Min", minutes))
	}
	return strings.Join(parts, " ")
}
