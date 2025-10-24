package handler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/response"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type AdminPOSHandler struct {
	requests repository.PosRequestRepository
	devices  repository.PosDeviceRepository
}

func NewAdminPOSHandler(reqs repository.PosRequestRepository, devices repository.PosDeviceRepository) *AdminPOSHandler {
	return &AdminPOSHandler{requests: reqs, devices: devices}
}

// ListRequests godoc
// @Summary List POS requests
// @Tags admin-pos
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status filter"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/requests [get]
func (h *AdminPOSHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	var status *domain.PosRequestStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.PosRequestStatus(s)
		status = &st
	}
	items, err := h.requests.List(r.Context(), status)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to list requests")
		return
	}
	type dto struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Model     string `json:"model"`
		OS        string `json:"os"`
		Status    string `json:"status"`
		CreatedAt any    `json:"createdAt"`
		ExpiresAt any    `json:"expiresAt"`
	}
	out := make([]dto, 0, len(items))
	for _, it := range items {
		out = append(out, dto{ID: it.ID.Hex(), Name: it.Name, Model: it.Model, OS: it.OS, Status: string(it.Status), CreatedAt: it.CreatedAt, ExpiresAt: it.ExpiresAt})
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": out})
}

// ListDevices godoc
// @Summary List POS devices
// @Tags admin-pos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/devices [get]
func (h *AdminPOSHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	items, err := h.devices.List(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to list devices")
		return
	}
	type dto struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Token       string  `json:"token"`
		Approved    bool    `json:"approved"`
		ApprovedAt  any     `json:"approvedAt,omitempty"`
		CardCapable *bool   `json:"cardCapable,omitempty"`
		PrinterMAC  *string `json:"printerMac,omitempty"`
		PrinterUUID *string `json:"printerUuid,omitempty"`
	}
	out := make([]dto, 0, len(items))
	for _, it := range items {
		out = append(out, dto{ID: it.ID.Hex(), Name: it.Name, Token: it.DeviceToken, Approved: it.Approved, ApprovedAt: it.ApprovedAt, CardCapable: it.CardCapable, PrinterMAC: it.PrinterMAC, PrinterUUID: it.PrinterUUID})
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": out})
}

// Approve godoc
// @Summary Approve POS request
// @Tags admin-pos
// @Security BearerAuth
// @Param id path string true "Request ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/requests/{id}/approve [post]
func (h *AdminPOSHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	req, err := h.requests.FindByID(r.Context(), oid)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	// Create device on approval if it doesn't exist yet, then mark approved
	if _, err := h.devices.UpsertPendingByToken(r.Context(), req.Name, req.DeviceToken); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "device create failed")
		return
	}
	if err := h.devices.ApproveByToken(r.Context(), req.DeviceToken); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "approve failed")
		return
	}
	var decidedBy *primitive.ObjectID
	if claims, ok := middleware.GetUserFromContext(r.Context()); ok && claims != nil && claims.Subject != "" {
		if u, err := primitive.ObjectIDFromHex(claims.Subject); err == nil {
			decidedBy = &u
		}
	}
	if err := h.requests.UpdateStatus(r.Context(), oid, domain.PosRequestStatusApproved, decidedBy, time.Now().UTC()); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Ack{Message: "approved"})
}

// Reject godoc
// @Summary Reject POS request
// @Tags admin-pos
// @Security BearerAuth
// @Param id path string true "Request ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/requests/{id}/reject [post]
func (h *AdminPOSHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if _, err := h.requests.FindByID(r.Context(), oid); err != nil {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	var decidedBy *primitive.ObjectID
	if claims, ok := middleware.GetUserFromContext(r.Context()); ok && claims != nil && claims.Subject != "" {
		if u, err := primitive.ObjectIDFromHex(claims.Subject); err == nil {
			decidedBy = &u
		}
	}
	if err := h.requests.UpdateStatus(r.Context(), oid, domain.PosRequestStatusRejected, decidedBy, time.Now().UTC()); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Ack{Message: "rejected"})
}

// PatchConfig godoc
// @Summary Update POS device config
// @Tags admin-pos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/devices/{id}/config [patch]
func (h *AdminPOSHandler) PatchConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		CardCapable *bool   `json:"cardCapable"`
		PrinterMAC  *string `json:"printerMac"`
		PrinterUUID *string `json:"printerUuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.devices.UpdateConfig(r.Context(), oid, body.CardCapable, body.PrinterMAC, body.PrinterUUID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Ack{Message: "updated"})
}
