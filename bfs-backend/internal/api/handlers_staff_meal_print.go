package api

import (
	"fmt"
	"net/http"
	"strconv"

	"backend/internal/pdf"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	printSlipDefaultCount = 30
	printSlipMinCount     = 1
	printSlipMaxCount     = 500
)

// PrintStaffMealSlips streams a PDF of printable QR slips for a campaign.
// GET /v1/staff-meals/{campaignId}/print.pdf?count=30
func (h *Handlers) PrintStaffMealSlips(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "campaignId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}

	count := printSlipDefaultCount
	if v := r.URL.Query().Get("count"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < printSlipMinCount || n > printSlipMaxCount {
			writeError(w, http.StatusBadRequest, "invalid_count",
				fmt.Sprintf("count must be between %d and %d", printSlipMinCount, printSlipMaxCount))
			return
		}
		count = n
	}

	detail, err := h.volunteers.GetCampaign(r.Context(), id)
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}

	products := make([]pdf.SlipProduct, 0, len(detail.Products))
	for _, p := range detail.Products {
		products = append(products, pdf.SlipProduct{
			Name:     p.ProductName,
			Quantity: p.Quantity,
		})
	}

	body, err := pdf.RenderStaffMealSlips(pdf.SlipInput{
		CampaignName: detail.Campaign.Name,
		QRPayload:    service.BuildQRPayload(detail.Campaign.ClaimToken),
		Products:     products,
		Count:        count,
	})
	if err != nil {
		h.logger.Error("render staff meal slips", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "render_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="staff-meal-%s.pdf"`, detail.Campaign.ID.String()))
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	_, _ = w.Write(body)
}
