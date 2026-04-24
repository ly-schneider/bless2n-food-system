package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/volunteercampaign"
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	volunteerSessionCookie = "hsclaim_session"
)

type volunteerCampaignProductPayload struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
}

type createVolunteerCampaignRequest struct {
	Name           string                            `json:"name"`
	ValidFrom      *time.Time                        `json:"validFrom,omitempty"`
	ValidUntil     *time.Time                        `json:"validUntil,omitempty"`
	Products       []volunteerCampaignProductPayload `json:"products"`
	MaxRedemptions int                               `json:"maxRedemptions"`
}

type updateVolunteerCampaignRequest struct {
	Name           string                            `json:"name"`
	ValidFrom      *time.Time                        `json:"validFrom,omitempty"`
	ValidUntil     *time.Time                        `json:"validUntil,omitempty"`
	Status         string                            `json:"status"`
	Products       []volunteerCampaignProductPayload `json:"products,omitempty"`
	MaxRedemptions *int                              `json:"maxRedemptions,omitempty"`
}

type verifyAccessRequest struct {
	Code string `json:"code"`
}

type adminCampaignResponse struct {
	ID              uuid.UUID  `json:"id"`
	ClaimToken      uuid.UUID  `json:"claimToken"`
	Name            string     `json:"name"`
	AccessCode      string     `json:"accessCode"`
	ValidFrom       *time.Time `json:"validFrom,omitempty"`
	ValidUntil      *time.Time `json:"validUntil,omitempty"`
	Status          string     `json:"status"`
	MaxRedemptions  int        `json:"maxRedemptions"`
	RedemptionCount int        `json:"redemptionCount"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type adminCampaignDetailResponse struct {
	adminCampaignResponse
	Products    []adminCampaignProductItem    `json:"products"`
	Redemptions []adminCampaignRedemptionItem `json:"redemptions"`
}

type adminCampaignProductItem struct {
	ProductID    uuid.UUID `json:"productId"`
	ProductName  string    `json:"productName"`
	ProductImage *string   `json:"productImage,omitempty"`
	Quantity     int       `json:"quantity"`
}

type adminCampaignRedemptionItem struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"orderId"`
	CreatedAt time.Time `json:"createdAt"`
}

type claimCampaignPublic struct {
	Name       string     `json:"name"`
	ValidFrom  *time.Time `json:"validFrom,omitempty"`
	ValidUntil *time.Time `json:"validUntil,omitempty"`
	Status     string     `json:"status"`
}

type claimCampaignResponse struct {
	Campaign  claimCampaignPublic        `json:"campaign"`
	Products  []adminCampaignProductItem `json:"products"`
	QRPayload string                     `json:"qrPayload"`
}

// CreateVolunteerCampaign (POST /v1/staff-meals)
func (h *Handlers) CreateVolunteerCampaign(w http.ResponseWriter, r *http.Request) {
	var req createVolunteerCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	products := make([]repository.VolunteerCampaignProductInput, 0, len(req.Products))
	for _, p := range req.Products {
		products = append(products, repository.VolunteerCampaignProductInput{
			ProductID: p.ProductID,
			Quantity:  p.Quantity,
		})
	}
	campaign, err := h.volunteers.CreateCampaign(r.Context(), service.CreateVolunteerCampaignInput{
		Name:           req.Name,
		ValidFrom:      req.ValidFrom,
		ValidUntil:     req.ValidUntil,
		Products:       products,
		MaxRedemptions: req.MaxRedemptions,
	})
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, campaignToAdminResponse(campaign))
}

// ListVolunteerCampaigns (GET /v1/staff-meals)
func (h *Handlers) ListVolunteerCampaigns(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.volunteers.ListCampaigns(r.Context())
	if err != nil {
		h.logger.Error("list volunteer campaigns", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]adminCampaignResponse, 0, len(summaries))
	for _, s := range summaries {
		items = append(items, campaignToAdminResponse(s.Campaign))
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

// GetVolunteerCampaign (GET /v1/staff-meals/{campaignId})
func (h *Handlers) GetVolunteerCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "campaignId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}
	detail, err := h.volunteers.GetCampaign(r.Context(), id)
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	products := make([]adminCampaignProductItem, 0, len(detail.Products))
	for _, cp := range detail.Products {
		products = append(products, adminCampaignProductItem{
			ProductID:    cp.ProductID,
			ProductName:  cp.ProductName,
			ProductImage: cp.ProductImage,
			Quantity:     cp.Quantity,
		})
	}
	redemptions := make([]adminCampaignRedemptionItem, 0, len(detail.Redemptions))
	for _, rv := range detail.Redemptions {
		redemptions = append(redemptions, adminCampaignRedemptionItem{
			ID:        rv.ID,
			OrderID:   rv.OrderID,
			CreatedAt: rv.CreatedAt,
		})
	}
	resp := adminCampaignDetailResponse{
		adminCampaignResponse: campaignToAdminResponse(detail.Campaign),
		Products:              products,
		Redemptions:           redemptions,
	}
	response.WriteJSON(w, http.StatusOK, resp)
}

// UpdateVolunteerCampaign (PATCH /v1/staff-meals/{campaignId})
func (h *Handlers) UpdateVolunteerCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "campaignId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}
	var req updateVolunteerCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	status := volunteercampaign.StatusActive
	if req.Status != "" {
		if err := volunteercampaign.StatusValidator(volunteercampaign.Status(req.Status)); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_status", err.Error())
			return
		}
		status = volunteercampaign.Status(req.Status)
	}
	products := make([]repository.VolunteerCampaignProductInput, 0, len(req.Products))
	for _, p := range req.Products {
		products = append(products, repository.VolunteerCampaignProductInput{
			ProductID: p.ProductID,
			Quantity:  p.Quantity,
		})
	}
	campaign, err := h.volunteers.UpdateCampaign(r.Context(), id, service.UpdateVolunteerCampaignInput{
		Name:           req.Name,
		ValidFrom:      req.ValidFrom,
		ValidUntil:     req.ValidUntil,
		Status:         status,
		Products:       products,
		MaxRedemptions: req.MaxRedemptions,
	})
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, campaignToAdminResponse(campaign))
}

// EndVolunteerCampaign (POST /v1/staff-meals/{campaignId}/end)
func (h *Handlers) EndVolunteerCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "campaignId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}
	if err := h.volunteers.EndCampaign(r.Context(), id); err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	response.WriteNoContent(w)
}

// RotateVolunteerCampaignToken (POST /v1/staff-meals/{campaignId}/rotate-token)
func (h *Handlers) RotateVolunteerCampaignToken(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "campaignId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}
	newToken, err := h.volunteers.RotateClaimToken(r.Context(), id)
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"claimToken": newToken})
}

// VerifyClaimAccess (POST /v1/claim/{token}/auth)
func (h *Handlers) VerifyClaimAccess(w http.ResponseWriter, r *http.Request) {
	token, err := uuid.Parse(chi.URLParam(r, "token"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token", err.Error())
		return
	}
	var req verifyAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	existing := readSessionCookie(r)
	sessionID, err := h.volunteers.VerifyAccess(r.Context(), token, req.Code)
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	if existing != "" {
		sessionID = existing
	}
	writeSessionCookie(w, r, sessionID)
	response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GetClaimCampaign (GET /v1/claim/{token})
func (h *Handlers) GetClaimCampaign(w http.ResponseWriter, r *http.Request) {
	token, err := uuid.Parse(chi.URLParam(r, "token"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token", err.Error())
		return
	}
	if readSessionCookie(r) == "" {
		writeError(w, http.StatusUnauthorized, "auth_required", "Access code required.")
		return
	}
	view, err := h.volunteers.GetClaimView(r.Context(), token)
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	products := make([]adminCampaignProductItem, 0, len(view.Products))
	for _, p := range view.Products {
		products = append(products, adminCampaignProductItem{
			ProductID:    p.ProductID,
			ProductName:  p.ProductName,
			ProductImage: p.ProductImage,
			Quantity:     p.Quantity,
		})
	}
	response.WriteJSON(w, http.StatusOK, claimCampaignResponse{
		Campaign: claimCampaignPublic{
			Name:       view.Campaign.Name,
			ValidFrom:  view.Campaign.ValidFrom,
			ValidUntil: view.Campaign.ValidUntil,
			Status:     string(view.Campaign.Status),
		},
		Products:  products,
		QRPayload: view.QRPayload,
	})
}

func (h *Handlers) writeVolunteerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrVolunteerCampaignNotFound):
		writeError(w, http.StatusNotFound, "campaign_not_found", "The volunteer campaign does not exist.")
	case errors.Is(err, service.ErrVolunteerCampaignInactive):
		writeError(w, http.StatusGone, "campaign_inactive", "This volunteer campaign is no longer active.")
	case errors.Is(err, service.ErrVolunteerCampaignOutsideValid):
		writeError(w, http.StatusGone, "campaign_outside_validity", "This volunteer campaign is outside its validity window.")
	case errors.Is(err, service.ErrVolunteerAccessCodeInvalid):
		writeError(w, http.StatusUnauthorized, "invalid_access_code", "The access code is incorrect.")
	case errors.Is(err, service.ErrVolunteerCampaignInvalidAccess):
		writeError(w, http.StatusBadRequest, "invalid_access_code_format", "Access code must be 4 digits.")
	case errors.Is(err, service.ErrVolunteerCampaignHasNoProducts):
		writeError(w, http.StatusBadRequest, "no_products", "At least one product is required.")
	case errors.Is(err, service.ErrVolunteerMaxRedemptionsReached):
		writeError(w, http.StatusConflict, "max_redemptions_reached", "This campaign has reached its maximum number of redemptions.")
	case errors.Is(err, service.ErrVolunteerMaxBelowCount):
		writeError(w, http.StatusConflict, "max_redemptions_below_count", "New maximum cannot be lower than the current redemption count.")
	case errors.Is(err, service.ErrVolunteerStationNoMatchingProducts):
		writeError(w, http.StatusConflict, "station_no_matching_products", "Diese Station bietet keine Artikel aus diesem Helfer-Essen an.")
	default:
		h.logger.Error("volunteer error", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
	}
}

func campaignToAdminResponse(c *ent.VolunteerCampaign) adminCampaignResponse {
	return adminCampaignResponse{
		ID:              c.ID,
		ClaimToken:      c.ClaimToken,
		Name:            c.Name,
		AccessCode:      c.AccessCode,
		ValidFrom:       c.ValidFrom,
		ValidUntil:      c.ValidUntil,
		Status:          string(c.Status),
		MaxRedemptions:  c.MaxRedemptions,
		RedemptionCount: c.RedemptionCount,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

func readSessionCookie(r *http.Request) string {
	c, err := r.Cookie(volunteerSessionCookie)
	if err != nil {
		return ""
	}
	return c.Value
}

func writeSessionCookie(w http.ResponseWriter, r *http.Request, sessionID string) {
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     volunteerSessionCookie,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})
}
