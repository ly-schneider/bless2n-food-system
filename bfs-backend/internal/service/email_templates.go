package service

import (
	"fmt"
	"time"
)

// InviteEmailData contains the data for rendering an admin invite email.
type InviteEmailData struct {
	Brand     string
	InviteURL string
	ExpiresAt string
}

// formatExpiry formats a time.Time in German Swiss locale.
func formatExpiry(t time.Time) string {
	// Format as "2. Januar 2025, 14:30"
	loc, _ := time.LoadLocation("Europe/Zurich")
	t = t.In(loc)

	months := []string{
		"Januar", "Februar", "März", "April", "Mai", "Juni",
		"Juli", "August", "September", "Oktober", "November", "Dezember",
	}
	month := months[t.Month()-1]

	return fmt.Sprintf("%d. %s %d, %02d:%02d", t.Day(), month, t.Year(), t.Hour(), t.Minute())
}

// renderInviteHTML generates the HTML version of the admin invite email.
func renderInviteHTML(data InviteEmailData) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="de" dir="ltr"
      xmlns:v="urn:schemas-microsoft-com:vml"
      xmlns:o="urn:schemas-microsoft-com:office:office">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width">
    <meta name="x-apple-disable-message-reformatting">
    <meta name="format-detection" content="telephone=no,address=no,email=no,date=no,url=no">
    <meta name="color-scheme" content="light">
    <meta name="supported-color-schemes" content="light">
    <title>%s Admin-Einladung</title>
    <!--[if mso]>
      <noscript>
        <xml>
          <o:OfficeDocumentSettings>
            <o:PixelsPerInch>96</o:PixelsPerInch>
          </o:OfficeDocumentSettings>
        </xml>
      </noscript>
      <style>
        body, table, td, p, a, span { font-family: Arial, sans-serif !important; }
      </style>
    <![endif]-->
    <style>
      @media (max-width:600px){
        .container { width:100%% !important; }
        .px { padding-left:16px !important; padding-right:16px !important; }
      }
    </style>
  </head>
  <body style="margin:0;padding:0;">
    <div style="display:none;max-height:0;overflow:hidden;mso-hide:all;color:transparent;opacity:0;">
      Du wurdest eingeladen, Admin bei %s zu werden &nbsp;&#8205;&nbsp;&#8205;&nbsp;&#8205;
    </div>
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0"
           bgcolor="#E9E7E6" style="background-color:#E9E7E6;">
      <tr>
        <td align="center" style="padding:24px;">
          <table role="presentation" width="560" cellpadding="0" cellspacing="0" border="0"
                 class="container"
                 bgcolor="#FDFDFD"
                 style="width:560px;max-width:560px;background-color:#FDFDFD;border:1px solid #D7D7D7;border-radius:11px;">
            <tr>
              <td class="px" style="padding:20px 24px 8px 24px;">
                <table role="presentation" cellpadding="0" cellspacing="0" border="0">
                  <tr>
                    <td style="font:700 14px/1.2 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#000000;">
                      %s
                    </td>
                  </tr>
                </table>
              </td>
            </tr>
            <tr>
              <td class="px" style="padding:8px 24px 24px 24px;">
                <p style="margin:0 0 8px 0;font:600 18px/1.3 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#000000;">
                  Admin-Einladung
                </p>
                <p style="margin:0 0 16px 0;font:13px/1.5 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                  Du wurdest eingeladen, als Administrator bei %s beizutreten.
                  Klicke auf den Button unten, um die Einladung anzunehmen.
                </p>
                <table role="presentation" cellpadding="0" cellspacing="0" border="0" style="margin:16px 0;">
                  <tr>
                    <td align="center" bgcolor="#000000" style="border-radius:7px;">
                      <a href="%s" target="_blank" style="display:inline-block;padding:14px 28px;font:600 14px/1 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#FFFFFF;text-decoration:none;">
                        Einladung annehmen
                      </a>
                    </td>
                  </tr>
                </table>
                <p style="margin:16px 0 0 0;font:12px/1.5 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                  Diese Einladung ist gueltig bis: %s
                </p>
                <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin:16px 0;">
                  <tr><td height="1" style="line-height:1px;font-size:1px;background:#D7D7D7;">&nbsp;</td></tr>
                </table>
                <p style="margin:0;font:12px/1.5 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                  Falls der Button nicht funktioniert, kopiere diesen Link in deinen Browser:
                </p>
                <p style="margin:8px 0 0 0;font:11px/1.5 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,'Liberation Mono','Courier New',monospace;color:#7B7B7B;word-break:break-all;">
                  %s
                </p>
              </td>
            </tr>
            <tr>
              <td class="px" style="padding:16px 24px 24px 24px;">
                <p style="margin:0;font:12px/1.5 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                  Wenn du diese Einladung nicht erwartet hast, kannst du diese E-Mail ignorieren.
                </p>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`, data.Brand, data.Brand, data.Brand, data.Brand, data.InviteURL, data.ExpiresAt, data.InviteURL)
}

// renderInviteText generates the plain text version of the admin invite email.
func renderInviteText(data InviteEmailData) string {
	return fmt.Sprintf(`%s Admin-Einladung

Du wurdest eingeladen, als Administrator bei %s beizutreten.

Einladung annehmen:
%s

Diese Einladung ist gueltig bis: %s

Wenn du diese Einladung nicht erwartet hast, kannst du diese E-Mail ignorieren.
`, data.Brand, data.Brand, data.InviteURL, data.ExpiresAt)
}

// OTPType represents the type of OTP being sent.
type OTPType string

const (
	OTPTypeSignIn            OTPType = "sign-in"
	OTPTypeEmailVerification OTPType = "email-verification"
	OTPTypeForgetPassword    OTPType = "forget-password"
)

// OTPEmailData contains the data for rendering an OTP email.
type OTPEmailData struct {
	Brand       string
	Code        string
	CodeTTL     string
	SupportNote string
}

// friendlyTTL formats duration in a human-friendly German format.
func friendlyTTL(seconds int) string {
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24

	var parts []string

	if days > 0 {
		if days == 1 {
			parts = append(parts, "1 Tag")
		} else {
			parts = append(parts, fmt.Sprintf("%d Tagen", days))
		}
	}

	remainingHours := hours % 24
	if remainingHours > 0 {
		parts = append(parts, fmt.Sprintf("%d Std", remainingHours))
	}

	remainingMinutes := minutes % 60
	if remainingMinutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d Min", remainingMinutes))
	}

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}

// getOTPSubject returns the email subject based on OTP type.
func getOTPSubject(otpType OTPType) string {
	switch otpType {
	case OTPTypeSignIn:
		return "Dein Anmeldecode"
	case OTPTypeForgetPassword:
		return "Passwort zurücksetzen"
	default:
		return "Dein Verifizierungscode"
	}
}

// renderOTPHTML generates the HTML version of the OTP email.
func renderOTPHTML(data OTPEmailData) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="de" dir="ltr"
      xmlns:v="urn:schemas-microsoft-com:vml"
      xmlns:o="urn:schemas-microsoft-com:office:office">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width">
    <meta name="x-apple-disable-message-reformatting">
    <meta name="format-detection" content="telephone=no,address=no,email=no,date=no,url=no">
    <meta name="color-scheme" content="light">
    <meta name="supported-color-schemes" content="light">
    <title>%s Anmeldung</title>
    <!--[if mso]>
      <noscript>
        <xml>
          <o:OfficeDocumentSettings>
            <o:PixelsPerInch>96</o:PixelsPerInch>
          </o:OfficeDocumentSettings>
        </xml>
      </noscript>
      <style>
        body, table, td, p, a, span { font-family: Arial, sans-serif !important; }
      </style>
    <![endif]-->
    <style>
      @media (max-width:600px){
        .container { width:100%% !important; }
        .px { padding-left:16px !important; padding-right:16px !important; }
      }
    </style>
  </head>
  <body style="margin:0;padding:0;">
    <div style="display:none;max-height:0;overflow:hidden;mso-hide:all;color:transparent;opacity:0;">
      Dein einmaliger Code für %s: %s (läuft in %s ab) &nbsp;&#8205;&nbsp;&#8205;&nbsp;&#8205;
    </div>
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0"
           bgcolor="#E9E7E6" style="background-color:#E9E7E6;">
      <tr>
        <td align="center" style="padding:24px;">
          <table role="presentation" width="560" cellpadding="0" cellspacing="0" border="0"
                 class="container"
                 bgcolor="#FDFDFD"
                 style="width:560px;max-width:560px;background-color:#FDFDFD;border:1px solid #D7D7D7;border-radius:11px;">
            <tr>
              <td class="px" style="padding:20px 24px 8px 24px;">
                <table role="presentation" cellpadding="0" cellspacing="0" border="0">
                  <tr>
                    <td style="font:700 14px/1.2 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#000000;">
                      %s
                    </td>
                  </tr>
                </table>
              </td>
            </tr>
            <tr>
              <td class="px" style="padding:8px 24px 24px 24px;">
                <p style="margin:0 0 8px 0;font:600 18px/1.3 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#000000;">
                  Anmelden mit Code
                </p>
                <p style="margin:0 0 16px 0;font:13px/1.5 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                  Gib diesen einmaligen 6-stelligen Code im Anmeldefenster ein:
                </p>
                <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0"
                       style="background:#FFFFFF;border:1px solid #D7D7D7;border-radius:7px;">
                  <tr>
                    <td align="center" style="padding:16px 12px;">
                      <span style="display:inline-block;font:700 28px/1.2 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,'Liberation Mono','Courier New',monospace;color:#000000;letter-spacing:6px;">
                        %s
                      </span>
                    </td>
                  </tr>
                </table>
                <p style="margin:16px 0 0 0;font:12px/1.5 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                  Der Code läuft in %s ab und kann nur einmal verwendet werden.
                </p>
                <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin:16px 0;">
                  <tr><td height="1" style="line-height:1px;font-size:1px;background:#D7D7D7;">&nbsp;</td></tr>
                </table>
                <p style="margin:0;font:12px/1.5 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                  %s
                </p>
              </td>
            </tr>
            <tr>
              <td class="px" style="padding:16px 24px 24px 24px;">
                <p style="margin:0;font:12px/1.5 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                  Wenn du dies nicht angefordert hast, kannst du diese E-Mail ignorieren.
                </p>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`, data.Brand, data.Brand, data.Code, data.CodeTTL, data.Brand, data.Code, data.CodeTTL, data.SupportNote)
}

// renderOTPText generates the plain text version of the OTP email.
func renderOTPText(data OTPEmailData) string {
	return fmt.Sprintf(`%s Anmeldung

Dein Code (läuft in %s ab):
  %s

%s

Wenn du dies nicht angefordert hast, kannst du diese E-Mail ignorieren.
`, data.Brand, data.CodeTTL, data.Code, data.SupportNote)
}
