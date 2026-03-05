package service

import (
	"fmt"
	"strings"
	"time"
)

type ReceiptLineItem struct {
	Title    string
	Quantity int
	Cents    int64
	Children []ReceiptLineItem
}

type ReceiptEmailData struct {
	Brand      string
	OrderID    string
	OrderURL   string
	OrderDate  string
	Items      []ReceiptLineItem
	TotalCents int64
	Method     string
}

func formatCHF(cents int64) string {
	whole := cents / 100
	frac := cents % 100
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("CHF %d.%02d", whole, frac)
}

func formatOrderDate(t time.Time) string {
	loc, _ := time.LoadLocation("Europe/Zurich")
	t = t.In(loc)
	months := []string{
		"Januar", "Februar", "März", "April", "Mai", "Juni",
		"Juli", "August", "September", "Oktober", "November", "Dezember",
	}
	return fmt.Sprintf("%d. %s %d, %02d:%02d", t.Day(), months[t.Month()-1], t.Year(), t.Hour(), t.Minute())
}

func renderReceiptHTML(data ReceiptEmailData) string {
	var itemRows strings.Builder
	for _, item := range data.Items {
		lineTotal := item.Cents * int64(item.Quantity)
		itemRows.WriteString(fmt.Sprintf(`
                    <tr>
                      <td style="padding:8px 0;font:14px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#000000;border-bottom:1px solid #EEEEEE;">
                        %s &times; %d
                      </td>
                      <td align="right" style="padding:8px 0;font:14px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#000000;border-bottom:1px solid #EEEEEE;">
                        %s
                      </td>
                    </tr>`, escHTML(item.Title), item.Quantity, formatCHF(lineTotal)))
		for _, child := range item.Children {
			itemRows.WriteString(fmt.Sprintf(`
                    <tr>
                      <td colspan="2" style="padding:2px 0 2px 16px;font:12px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;border-bottom:1px solid #EEEEEE;">
                        %s
                      </td>
                    </tr>`, escHTML(child.Title)))
		}
	}

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
    <title>%s Quittung</title>
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
      Deine Quittung von %s — %s &nbsp;&#8205;&nbsp;&#8205;&nbsp;&#8205;
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
                <p style="margin:0 0 4px 0;font:600 18px/1.3 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#000000;">
                  Deine Quittung
                </p>
                <p style="margin:0 0 20px 0;font:13px/1.5 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                  Vielen Dank für deine Bestellung!
                </p>

                <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0"
                       style="background:#FAFAFA;border:1px solid #EEEEEE;border-radius:7px;">
                  <tr>
                    <td style="padding:12px 16px;">
                      <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0">
                        <tr>
                          <td style="font:12px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;padding-bottom:2px;">
                            Bestellnr.
                          </td>
                        </tr>
                        <tr>
                          <td style="font:10px/1.3 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;color:#999999;padding-bottom:8px;word-break:break-all;">
                            %s
                          </td>
                        </tr>
                        <tr>
                          <td style="font:12px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                            Datum
                          </td>
                          <td align="right" style="font:12px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                            %s
                          </td>
                        </tr>
                        <tr>
                          <td style="font:12px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                            Zahlungsart
                          </td>
                          <td align="right" style="font:12px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                            %s
                          </td>
                        </tr>
                      </table>
                    </td>
                  </tr>
                </table>

                <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin-top:16px;">
%s
                    <tr>
                      <td style="padding:12px 0 0 0;font:600 15px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#000000;">
                        Total
                      </td>
                      <td align="right" style="padding:12px 0 0 0;font:600 15px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#000000;">
                        %s
                      </td>
                    </tr>
                </table>

                <p style="margin:8px 0 0 0;font:11px/1.4 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                  Alle Preise in CHF inkl. MwSt.
                </p>

                <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%" style="margin:20px 0 0 0;">
                  <tr>
                    <td align="center" bgcolor="#000000" style="border-radius:11px;">
                      <a href="%s" target="_blank" style="display:block;padding:16px 28px;font:600 16px/1 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#FFFFFF;text-decoration:none;text-align:center;">
                        Bestellung anzeigen &amp; QR-Code
                      </a>
                    </td>
                  </tr>
                </table>

                <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin:24px 0 0 0;">
                  <tr><td height="1" style="line-height:1px;font-size:1px;background:#D7D7D7;">&nbsp;</td></tr>
                </table>

                <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin:16px 0 0 0;">
                  <tr>
                    <td style="font:12px/1.6 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#7B7B7B;">
                      <strong style="color:#000000;">BlessThun</strong><br>
                      Verein BlessThun<br>
                      Postfach, 3602 Thun<br>
                      info@blessthun.ch
                    </td>
                  </tr>
                </table>
                <p style="margin:12px 0 0 0;font:11px/1.5 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#999999;">
                  Dies ist eine automatisch generierte Quittung.
                  Bitte bewahre diese E-Mail als Zahlungsbeleg auf.
                </p>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`,
		data.Brand,
		data.Brand, formatCHF(data.TotalCents),
		data.Brand,
		escHTML(data.OrderID),
		escHTML(data.OrderDate),
		escHTML(data.Method),
		itemRows.String(),
		formatCHF(data.TotalCents),
		escHTML(data.OrderURL),
	)
}

func renderReceiptText(data ReceiptEmailData) string {
	var lines strings.Builder
	for _, item := range data.Items {
		lineTotal := item.Cents * int64(item.Quantity)
		lines.WriteString(fmt.Sprintf("  %s x%d  %s\n", item.Title, item.Quantity, formatCHF(lineTotal)))
		for _, child := range item.Children {
			lines.WriteString(fmt.Sprintf("    - %s\n", child.Title))
		}
	}

	return fmt.Sprintf(`%s — Quittung

Vielen Dank für deine Bestellung!

Bestellnr.: %s
Datum: %s
Zahlungsart: %s

Artikel:
%s
Total: %s
Alle Preise in CHF inkl. MwSt.

Bestellung anzeigen & QR-Code:
%s

---
BlessThun
Verein BlessThun
Postfach, 3602 Thun
info@blessthun.ch

Dies ist eine automatisch generierte Quittung.
Bitte bewahre diese E-Mail als Zahlungsbeleg auf.
`, data.Brand, data.OrderID, data.OrderDate, data.Method, lines.String(), formatCHF(data.TotalCents), data.OrderURL)
}

func escHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
