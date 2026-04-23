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
	Name       string                            `json:"name"`
	AccessCode string                            `json:"accessCode"`
	ValidFrom  *time.Time                        `json:"validFrom,omitempty"`
	ValidUntil *time.Time                        `json:"validUntil,omitempty"`
	Products   []volunteerCampaignProductPayload `json:"products"`
	SlotCount  int                               `json:"slotCount"`
}

type updateVolunteerCampaignRequest struct {
	Name       string                            `json:"name"`
	AccessCode string                            `json:"accessCode"`
	ValidFrom  *time.Time                        `json:"validFrom,omitempty"`
	ValidUntil *time.Time                        `json:"validUntil,omitempty"`
	Status     string                            `json:"status"`
	Products   []volunteerCampaignProductPayload `json:"products,omitempty"`
}

type verifyAccessRequest struct {
	Code string `json:"code"`
}

type adminCampaignResponse struct {
	ID            uuid.UUID  `json:"id"`
	ClaimToken    uuid.UUID  `json:"claimToken"`
	Name          string     `json:"name"`
	AccessCode    string     `json:"accessCode"`
	ValidFrom     *time.Time `json:"validFrom,omitempty"`
	ValidUntil    *time.Time `json:"validUntil,omitempty"`
	Status        string     `json:"status"`
	TotalSlots    int        `json:"totalSlots"`
	RedeemedSlots int        `json:"redeemedSlots"`
	ReservedSlots int        `json:"reservedSlots"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type adminCampaignDetailResponse struct {
	adminCampaignResponse
	Products []adminCampaignProductItem `json:"products"`
	Slots    []adminCampaignSlotItem    `json:"slots"`
}

type adminCampaignProductItem struct {
	ProductID   uuid.UUID `json:"productId"`
	ProductName string    `json:"productName"`
	Quantity    int       `json:"quantity"`
}

type adminCampaignSlotItem struct {
	ID                uuid.UUID  `json:"id"`
	OrderID           uuid.UUID  `json:"orderId"`
	ReservedBySession *string    `json:"reservedBySession,omitempty"`
	ReservedUntil     *time.Time `json:"reservedUntil,omitempty"`
	IsRedeemed        bool       `json:"isRedeemed"`
	RedeemedAt        *time.Time `json:"redeemedAt,omitempty"`
}

type claimCampaignPublic struct {
	Name       string     `json:"name"`
	ValidFrom  *time.Time `json:"validFrom,omitempty"`
	ValidUntil *time.Time `json:"validUntil,omitempty"`
}

type claimListResponse struct {
	Campaign       claimCampaignPublic `json:"campaign"`
	TotalSlots     int                 `json:"totalSlots"`
	AvailableCount int                 `json:"availableCount"`
	Available      []claimSlotSummary  `json:"available"`
	ReservedByMe   []claimSlotSummary  `json:"reservedByMe"`
}

type claimSlotSummary struct {
	ID            uuid.UUID            `json:"id"`
	OrderID       uuid.UUID            `json:"orderId"`
	ReservedUntil *time.Time           `json:"reservedUntil,omitempty"`
	Lines         []claimSlotLineBrief `json:"lines"`
}

type claimSlotLineBrief struct {
	ProductName  string  `json:"productName"`
	ProductImage *string `json:"productImage,omitempty"`
	Quantity     int     `json:"quantity"`
}

type claimSlotDetail struct {
	ID            uuid.UUID            `json:"id"`
	OrderID       uuid.UUID            `json:"orderId"`
	ReservedUntil *time.Time           `json:"reservedUntil,omitempty"`
	IsRedeemed    bool                 `json:"isRedeemed"`
	RedeemedAt    *time.Time           `json:"redeemedAt,omitempty"`
	Lines         []claimSlotLineBrief `json:"lines"`
}

// CreateVolunteerCampaign (POST /v1/admin/volunteer-campaigns)
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
		Name:       req.Name,
		AccessCode: req.AccessCode,
		ValidFrom:  req.ValidFrom,
		ValidUntil: req.ValidUntil,
		Products:   products,
		SlotCount:  req.SlotCount,
	})
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	resp := campaignToAdminResponse(campaign, 0, 0, 0)
	resp.TotalSlots = req.SlotCount
	response.WriteJSON(w, http.StatusCreated, resp)
}

// ListVolunteerCampaigns (GET /v1/admin/volunteer-campaigns)
func (h *Handlers) ListVolunteerCampaigns(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.volunteers.ListCampaigns(r.Context())
	if err != nil {
		h.logger.Error("list volunteer campaigns", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]adminCampaignResponse, 0, len(summaries))
	for _, s := range summaries {
		items = append(items, campaignToAdminResponse(s.Campaign, s.TotalSlots, s.RedeemedSlots, s.ReservedSlots))
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

// GetVolunteerCampaign (GET /v1/admin/volunteer-campaigns/{campaignId})
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
	total, redeemed, reserved := 0, 0, 0
	now := time.Now()
	slotItems := make([]adminCampaignSlotItem, 0, len(detail.Slots))
	for _, slot := range detail.Slots {
		total++
		item := adminCampaignSlotItem{
			ID:            slot.ID,
			OrderID:       slot.OrderID,
			ReservedUntil: slot.ReservedUntil,
		}
		if slot.ReservedBySession != nil {
			masked := maskSession(*slot.ReservedBySession)
			item.ReservedBySession = &masked
		}
		redeemedAt, full := slotRedemptionTime(slot)
		if full {
			item.IsRedeemed = true
			item.RedeemedAt = redeemedAt
			redeemed++
		} else if slot.ReservedBySession != nil && slot.ReservedUntil != nil && slot.ReservedUntil.After(now) {
			reserved++
		}
		slotItems = append(slotItems, item)
	}
	products := make([]adminCampaignProductItem, 0, len(detail.Products))
	for _, cp := range detail.Products {
		p, _ := cp.Edges.ProductOrErr()
		name := ""
		if p != nil {
			name = p.Name
		}
		products = append(products, adminCampaignProductItem{
			ProductID:   cp.ProductID,
			ProductName: name,
			Quantity:    cp.Quantity,
		})
	}
	resp := adminCampaignDetailResponse{
		adminCampaignResponse: campaignToAdminResponse(detail.Campaign, total, redeemed, reserved),
		Products:              products,
		Slots:                 slotItems,
	}
	response.WriteJSON(w, http.StatusOK, resp)
}

// UpdateVolunteerCampaign (PATCH /v1/admin/volunteer-campaigns/{campaignId})
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
		Name:       req.Name,
		AccessCode: req.AccessCode,
		ValidFrom:  req.ValidFrom,
		ValidUntil: req.ValidUntil,
		Status:     status,
		Products:   products,
	})
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, campaignToAdminResponse(campaign, 0, 0, 0))
}

// EndVolunteerCampaign (POST /v1/admin/volunteer-campaigns/{campaignId}/end)
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

// RotateVolunteerCampaignToken (POST /v1/admin/volunteer-campaigns/{campaignId}/rotate-token)
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

// ListClaimSlots (GET /v1/claim/{token}/slots)
func (h *Handlers) ListClaimSlots(w http.ResponseWriter, r *http.Request) {
	token, err := uuid.Parse(chi.URLParam(r, "token"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token", err.Error())
		return
	}
	sessionID := readSessionCookie(r)
	if sessionID == "" {
		writeError(w, http.StatusUnauthorized, "auth_required", "Access code required.")
		return
	}
	view, err := h.volunteers.ListClaimSlots(r.Context(), token, sessionID)
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	resp := claimListResponse{
		Campaign: claimCampaignPublic{
			Name:       view.Campaign.Name,
			ValidFrom:  view.Campaign.ValidFrom,
			ValidUntil: view.Campaign.ValidUntil,
		},
		TotalSlots:     view.TotalSlots,
		AvailableCount: len(view.AvailableSlots),
		Available:      slotViewsToSummaries(view.AvailableSlots),
		ReservedByMe:   slotViewsToSummaries(view.ReservedByMe),
	}
	response.WriteJSON(w, http.StatusOK, resp)
}

// GetClaimSlot (GET /v1/claim/{token}/slots/{slotId})
func (h *Handlers) GetClaimSlot(w http.ResponseWriter, r *http.Request) {
	token, err := uuid.Parse(chi.URLParam(r, "token"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token", err.Error())
		return
	}
	slotID, err := uuid.Parse(chi.URLParam(r, "slotId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_slot_id", err.Error())
		return
	}
	sessionID := readSessionCookie(r)
	if sessionID == "" {
		writeError(w, http.StatusUnauthorized, "auth_required", "Access code required.")
		return
	}
	view, err := h.volunteers.GetClaimSlot(r.Context(), token, slotID, sessionID)
	if err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	lines := make([]claimSlotLineBrief, 0, len(view.Lines))
	for _, l := range view.Lines {
		lines = append(lines, claimSlotLineBrief{ProductName: l.ProductName, ProductImage: l.ProductImage, Quantity: l.Quantity})
	}
	response.WriteJSON(w, http.StatusOK, claimSlotDetail{
		ID:            view.Slot.ID,
		OrderID:       view.OrderID,
		ReservedUntil: view.ReservedUntil,
		IsRedeemed:    view.RedeemedAt != nil,
		RedeemedAt:    view.RedeemedAt,
		Lines:         lines,
	})
}

// ReserveClaimSlot (POST /v1/claim/{token}/slots/{slotId}/reserve)
func (h *Handlers) ReserveClaimSlot(w http.ResponseWriter, r *http.Request) {
	token, slotID, sessionID, ok := h.parseClaimParams(w, r)
	if !ok {
		return
	}
	if err := h.volunteers.ReserveSlot(r.Context(), token, slotID, sessionID); err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	response.WriteNoContent(w)
}

// ReleaseClaimSlot (POST /v1/claim/{token}/slots/{slotId}/release)
func (h *Handlers) ReleaseClaimSlot(w http.ResponseWriter, r *http.Request) {
	token, slotID, sessionID, ok := h.parseClaimParams(w, r)
	if !ok {
		return
	}
	if err := h.volunteers.ReleaseSlot(r.Context(), token, slotID, sessionID); err != nil {
		h.writeVolunteerError(w, err)
		return
	}
	response.WriteNoContent(w)
}

func (h *Handlers) parseClaimParams(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, string, bool) {
	token, err := uuid.Parse(chi.URLParam(r, "token"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token", err.Error())
		return uuid.Nil, uuid.Nil, "", false
	}
	slotID, err := uuid.Parse(chi.URLParam(r, "slotId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_slot_id", err.Error())
		return uuid.Nil, uuid.Nil, "", false
	}
	sessionID := readSessionCookie(r)
	if sessionID == "" {
		writeError(w, http.StatusUnauthorized, "auth_required", "Access code required.")
		return uuid.Nil, uuid.Nil, "", false
	}
	return token, slotID, sessionID, true
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
	case errors.Is(err, service.ErrVolunteerSlotNotAvailable):
		writeError(w, http.StatusConflict, "slot_not_available", "This slot is no longer available.")
	case errors.Is(err, service.ErrVolunteerSlotNotYours):
		writeError(w, http.StatusForbidden, "slot_not_yours", "You did not reserve this slot.")
	default:
		h.logger.Error("volunteer error", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
	}
}

func campaignToAdminResponse(c *ent.VolunteerCampaign, totalSlots, redeemedSlots, reservedSlots int) adminCampaignResponse {
	return adminCampaignResponse{
		ID:            c.ID,
		ClaimToken:    c.ClaimToken,
		Name:          c.Name,
		AccessCode:    c.AccessCode,
		ValidFrom:     c.ValidFrom,
		ValidUntil:    c.ValidUntil,
		Status:        string(c.Status),
		TotalSlots:    totalSlots,
		RedeemedSlots: redeemedSlots,
		ReservedSlots: reservedSlots,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

func slotViewsToSummaries(views []service.VolunteerSlotView) []claimSlotSummary {
	out := make([]claimSlotSummary, 0, len(views))
	for _, v := range views {
		lines := make([]claimSlotLineBrief, 0, len(v.Lines))
		for _, l := range v.Lines {
			if l.IsRedeemed {
				continue
			}
			lines = append(lines, claimSlotLineBrief{ProductName: l.ProductName, ProductImage: l.ProductImage, Quantity: l.Quantity})
		}
		out = append(out, claimSlotSummary{
			ID:            v.Slot.ID,
			OrderID:       v.OrderID,
			ReservedUntil: v.ReservedUntil,
			Lines:         lines,
		})
	}
	return out
}

func slotRedemptionTime(slot *ent.VolunteerSlot) (*time.Time, bool) {
	ord, err := slot.Edges.OrderOrErr()
	if err != nil || ord == nil {
		return nil, false
	}
	if len(ord.Edges.Lines) == 0 {
		return nil, false
	}
	var latest *time.Time
	for _, l := range ord.Edges.Lines {
		r, _ := l.Edges.RedemptionOrErr()
		if r == nil {
			return nil, false
		}
		t := r.RedeemedAt
		if latest == nil || t.After(*latest) {
			latest = &t
		}
	}
	return latest, true
}

func maskSession(s string) string {
	if len(s) <= 6 {
		return "***"
	}
	return s[:4] + "…" + s[len(s)-2:]
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
