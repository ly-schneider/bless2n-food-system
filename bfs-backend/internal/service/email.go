package service

import (
	"backend/internal/config"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type EmailService interface {
    SendLoginEmail(ctx context.Context, to, code, ip, ua string, codeTTL time.Duration) error
    PreviewLoginEmail(ctx context.Context, to, code, ip, ua string, codeTTL time.Duration) (subject, text, html string, err error)
    SendEmailChangeVerification(ctx context.Context, toNewEmail, code, ip, ua string, codeTTL time.Duration) error
    PreviewEmailChangeVerification(ctx context.Context, toNewEmail, code, ip, ua string, codeTTL time.Duration) (subject, text, html string, err error)
}

type emailService struct {
	cfg      config.Config
	htmlTmpl *template.Template
	textTmpl *template.Template
}

func NewEmailService(cfg config.Config) EmailService {
    return &emailService{
        cfg:      cfg,
        htmlTmpl: template.Must(template.New("login_html").Parse(loginEmailHTML)),
        textTmpl: template.Must(template.New("login_text").Parse(loginEmailText)),
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
		Brand:       "BlessThun",
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
	return e.send(ctx, to, subject, textBody.String(), htmlBody.String())
}

func (e *emailService) PreviewLoginEmail(ctx context.Context, to, code, ip, ua string, codeTTL time.Duration) (string, string, string, error) {
	data := loginData{
		Brand:       "BlessThun",
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
    Brand   string
    Code    string
    CodeTTL string
    NewEmail string
    IP      string
    UA      string
}

func (e *emailService) SendEmailChangeVerification(ctx context.Context, toNewEmail, code, ip, ua string, codeTTL time.Duration) error {
    data := emailChangeData{
        Brand:   "BlessThun",
        Code:    code,
        CodeTTL: friendlyTTL(codeTTL),
        NewEmail: toNewEmail,
        IP:      ip,
        UA:      ua,
    }
    htmlT := template.Must(template.New("email_change_html").Parse(emailChangeHTML))
    textT := template.Must(template.New("email_change_text").Parse(emailChangeText))
    var htmlBody, textBody strings.Builder
    if err := htmlT.Execute(&htmlBody, data); err != nil { return err }
    if err := textT.Execute(&textBody, data); err != nil { return err }
    return e.send(ctx, toNewEmail, "E‑Mail Änderung bestätigen", textBody.String(), htmlBody.String())
}

func (e *emailService) PreviewEmailChangeVerification(ctx context.Context, toNewEmail, code, ip, ua string, codeTTL time.Duration) (string, string, string, error) {
    data := emailChangeData{
        Brand:   "BlessThun",
        Code:    code,
        CodeTTL: friendlyTTL(codeTTL),
        NewEmail: toNewEmail,
        IP:      ip,
        UA:      ua,
    }
    htmlT := template.Must(template.New("email_change_html").Parse(emailChangeHTML))
    textT := template.Must(template.New("email_change_text").Parse(emailChangeText))
    var htmlBody, textBody strings.Builder
    if err := htmlT.Execute(&htmlBody, data); err != nil { return "","","", err }
    if err := textT.Execute(&textBody, data); err != nil { return "","","", err }
    return "E‑Mail Änderung bestätigen", textBody.String(), htmlBody.String(), nil
}

func (e *emailService) send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	from := e.cfg.Smtp.From
	host := e.cfg.Smtp.Host
	port := e.cfg.Smtp.Port
	user := e.cfg.Smtp.Username
	pass := e.cfg.Smtp.Password
	tlsPolicy := strings.ToLower(e.cfg.Smtp.TLSPolicy)

	addr := fmt.Sprintf("%s:%s", host, port)
	// Build MIME message with multipart/alternative
	boundary := "bfs-mime-" + fmt.Sprint(time.Now().UnixNano())
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q", boundary),
	}
	var msg strings.Builder
	msg.WriteString(strings.Join(headers, "\r\n"))
	msg.WriteString("\r\n\r\n")
	// text part
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n")
	// html part
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n")
	// end
	msg.WriteString("--" + boundary + "--\r\n")

	auth := smtp.PlainAuth("", user, pass, host)

	// TLS policy
	switch tlsPolicy {
	case "tls":
		// Implicit TLS (port 465)
		conn, err := tlsDial(addr, host)
		if err != nil {
			return err
		}
		c, err := smtp.NewClient(conn, host)
		if err != nil {
			return err
		}
		defer func() { _ = c.Quit() }()
		if user != "" {
			if err := c.Auth(auth); err != nil {
				return err
			}
		}
		if err := c.Mail(from); err != nil {
			return err
		}
		if err := c.Rcpt(to); err != nil {
			return err
		}
		wc, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := wc.Write([]byte(msg.String())); err != nil {
			_ = wc.Close()
			return err
		}
		return wc.Close()
	case "none":
		// No TLS (dev/Mailpit)
		return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg.String()))
	default:
		// STARTTLS (default)
		c, err := smtp.Dial(addr)
		if err != nil {
			return err
		}
		defer func() { _ = c.Quit() }()
		// greet and STARTTLS
		if ok, _ := c.Extension("STARTTLS"); ok {
			cfg := &tls.Config{ServerName: host, InsecureSkipVerify: false}
			if err := c.StartTLS(cfg); err != nil {
				return err
			}
		}
		if user != "" {
			if err := c.Auth(auth); err != nil {
				return err
			}
		}
		if err := c.Mail(from); err != nil {
			return err
		}
		if err := c.Rcpt(to); err != nil {
			return err
		}
		wc, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := wc.Write([]byte(msg.String())); err != nil {
			_ = wc.Close()
			return err
		}
		return wc.Close()
	}
}

func tlsDial(addr, serverName string) (*tls.Conn, error) {
	d := &net.Dialer{Timeout: 10 * time.Second}
	return tls.DialWithDialer(d, "tcp", addr, &tls.Config{ServerName: serverName})
}

func friendlyTTL(d time.Duration) string {
	if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return d.String()
}
