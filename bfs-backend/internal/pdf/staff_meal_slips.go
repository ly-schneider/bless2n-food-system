package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	qrcode "github.com/skip2/go-qrcode"
)

type SlipProduct struct {
	Name     string
	Quantity int
}

type SlipInput struct {
	CampaignName string
	QRPayload    string
	Products     []SlipProduct
	Count        int
}

const (
	pageWidth             = 210.0
	pageHeight            = 297.0
	pageMarginMM          = 6.0
	slipColumns           = 5
	slipHGap              = 2.0
	slipVGap              = 2.0
	qrSizeMM              = 28.0
	paddingMM             = 2.5
	nameFontSize          = 8.0
	nameLineHeight        = 3.2
	instructionFontSize   = 5.5
	instructionLineHeight = 2.3
	itemFontSize          = 6.5
	itemLineHeight        = 2.6

	instructionText = "An der Station vorzeigen"
)

func RenderStaffMealSlips(in SlipInput) ([]byte, error) {
	if in.Count <= 0 {
		return nil, fmt.Errorf("count must be positive")
	}
	if strings.TrimSpace(in.QRPayload) == "" {
		return nil, fmt.Errorf("qr payload required")
	}

	qrPng, err := qrcode.Encode(in.QRPayload, qrcode.Medium, 512)
	if err != nil {
		return nil, fmt.Errorf("encode qr: %w", err)
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pageMarginMM, pageMarginMM, pageMarginMM)
	pdf.SetAutoPageBreak(false, pageMarginMM)
	pdf.SetTextColor(20, 20, 20)
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	pdf.RegisterImageOptionsReader("qr", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(qrPng))

	usableWidth := pageWidth - 2*pageMarginMM
	slipWidth := (usableWidth - float64(slipColumns-1)*slipHGap) / float64(slipColumns)

	campaignNameEnc := tr(in.CampaignName)
	instructionEnc := tr(instructionText)

	nameLines := wrapTextLines(campaignNameEnc, slipWidth-2*paddingMM, nameFontSize, pdf, "B")
	instructionLines := wrapTextLines(instructionEnc, slipWidth-2*paddingMM, instructionFontSize, pdf, "I")

	productLines := make([][]string, len(in.Products))
	totalItemLines := 0
	for i, p := range in.Products {
		line := tr(formatProductLine(p))
		wrapped := wrapTextLines(line, slipWidth-2*paddingMM, itemFontSize, pdf, "")
		productLines[i] = wrapped
		totalItemLines += len(wrapped)
	}

	slipHeight := paddingMM + qrSizeMM + 1.2 +
		float64(len(instructionLines))*instructionLineHeight + 1.5 +
		float64(len(nameLines))*nameLineHeight + 1.2 +
		float64(totalItemLines)*itemLineHeight + paddingMM

	maxRows := int((pageHeight - 2*pageMarginMM + slipVGap) / (slipHeight + slipVGap))
	if maxRows < 1 {
		maxRows = 1
	}
	perPage := maxRows * slipColumns

	idx := 0
	for idx < in.Count {
		pdf.AddPage()
		remaining := in.Count - idx
		pageSlips := perPage
		if remaining < pageSlips {
			pageSlips = remaining
		}
		for slot := 0; slot < pageSlips; slot++ {
			row := slot / slipColumns
			col := slot % slipColumns
			x := pageMarginMM + float64(col)*(slipWidth+slipHGap)
			y := pageMarginMM + float64(row)*(slipHeight+slipVGap)
			drawSlip(pdf, x, y, slipWidth, nameLines, instructionLines, productLines)
		}
		drawCutLines(pdf, slipWidth, slipHeight, pageSlips)
		idx += pageSlips
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func formatProductLine(p SlipProduct) string {
	qty := p.Quantity
	if qty < 1 {
		qty = 1
	}
	return fmt.Sprintf("%d\u00d7 %s", qty, p.Name)
}

func drawSlip(pdf *fpdf.Fpdf, x, y, w float64, nameLines, instructionLines []string, productLines [][]string) {
	qrX := x + (w-qrSizeMM)/2
	qrY := y + paddingMM
	pdf.ImageOptions("qr", qrX, qrY, qrSizeMM, qrSizeMM, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	textY := qrY + qrSizeMM + 1.2
	pdf.SetFont("Helvetica", "I", instructionFontSize)
	pdf.SetTextColor(110, 110, 110)
	for _, line := range instructionLines {
		pdf.SetXY(x+paddingMM, textY)
		pdf.CellFormat(w-2*paddingMM, instructionLineHeight, line, "", 0, "C", false, 0, "")
		textY += instructionLineHeight
	}
	pdf.SetTextColor(20, 20, 20)

	textY += 1.5
	pdf.SetFont("Helvetica", "B", nameFontSize)
	for _, line := range nameLines {
		pdf.SetXY(x+paddingMM, textY)
		pdf.CellFormat(w-2*paddingMM, nameLineHeight, line, "", 0, "C", false, 0, "")
		textY += nameLineHeight
	}

	textY += 1.2
	pdf.SetFont("Helvetica", "", itemFontSize)
	for _, wrapped := range productLines {
		for _, wl := range wrapped {
			pdf.SetXY(x+paddingMM, textY)
			pdf.CellFormat(w-2*paddingMM, itemLineHeight, wl, "", 0, "C", false, 0, "")
			textY += itemLineHeight
		}
	}
}

func drawCutLines(pdf *fpdf.Fpdf, slipWidth, slipHeight float64, slipsOnPage int) {
	pdf.SetDrawColor(160, 160, 160)
	pdf.SetLineWidth(0.1)
	pdf.SetDashPattern([]float64{0.8, 0.8}, 0)
	defer pdf.SetDashPattern([]float64{}, 0)

	rows := (slipsOnPage-1)/slipColumns + 1
	topEdge := pageMarginMM - slipVGap/2
	bottomEdge := pageMarginMM + float64(rows-1)*(slipHeight+slipVGap) + slipHeight + slipVGap/2

	for r := 0; r <= rows; r++ {
		var y float64
		switch r {
		case 0:
			y = topEdge
		case rows:
			y = bottomEdge
		default:
			y = pageMarginMM + float64(r)*(slipHeight+slipVGap) - slipVGap/2
		}
		pdf.Line(pageMarginMM-1, y, pageWidth-pageMarginMM+1, y)
	}
	for c := 1; c < slipColumns; c++ {
		x := pageMarginMM + float64(c)*slipWidth + float64(c-1)*slipHGap + slipHGap/2
		pdf.Line(x, topEdge, x, bottomEdge)
	}
}

func wrapTextLines(text string, maxWidth, fontSize float64, pdf *fpdf.Fpdf, style string) []string {
	pdf.SetFont("Helvetica", style, fontSize)
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}
	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		candidate := current + " " + w
		if pdf.GetStringWidth(candidate) <= maxWidth {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = w
	}
	lines = append(lines, current)
	return lines
}
