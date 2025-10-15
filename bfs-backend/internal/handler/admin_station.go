package handler

import (
    "backend/internal/domain"
    "backend/internal/middleware"
    "backend/internal/response"
    "backend/internal/repository"
    "net/http"
    "time"
    "encoding/json"

    "github.com/go-chi/chi/v5"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.uber.org/zap"
)

type AdminStationHandler struct {
    stationReqs repository.StationRequestRepository
    stations    repository.StationRepository
    stationProds repository.StationProductRepository
    products    repository.ProductRepository
    audit       repository.AuditRepository
    logger      *zap.Logger
}

func NewAdminStationHandler(reqs repository.StationRequestRepository, stations repository.StationRepository, stationProds repository.StationProductRepository, products repository.ProductRepository, audit repository.AuditRepository, logger *zap.Logger) *AdminStationHandler {
    return &AdminStationHandler{ stationReqs: reqs, stations: stations, stationProds: stationProds, products: products, audit: audit, logger: logger }
}

// ListRequests godoc
// @Summary List station requests
// @Tags admin-stations
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status filter"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/stations/requests [get]
func (h *AdminStationHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
    var status *domain.StationRequestStatus
    if s := r.URL.Query().Get("status"); s != "" {
        st := domain.StationRequestStatus(s)
        status = &st
    }
    items, err := h.stationReqs.List(r.Context(), status)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to list requests"); return }
    type dto struct { ID string `json:"id"`; Name string `json:"name"`; Model string `json:"model"`; OS string `json:"os"`; Status string `json:"status"`; CreatedAt any `json:"createdAt"`; ExpiresAt any `json:"expiresAt"` }
    out := make([]dto, 0, len(items))
    for _, it := range items { out = append(out, dto{ ID: it.ID.Hex(), Name: it.Name, Model: it.Model, OS: it.OS, Status: string(it.Status), CreatedAt: it.CreatedAt, ExpiresAt: it.ExpiresAt }) }
    response.WriteJSON(w, http.StatusOK, map[string]any{ "items": out })
}

// ListStations godoc
// @Summary List stations
// @Tags admin-stations
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/stations [get]
func (h *AdminStationHandler) ListStations(w http.ResponseWriter, r *http.Request) {
    items, err := h.stations.List(r.Context())
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to list stations"); return }
    type dto struct { ID string `json:"id"`; Name string `json:"name"`; DeviceKey string `json:"deviceKey"`; Approved bool `json:"approved"`; ApprovedAt any `json:"approvedAt,omitempty"`; CreatedAt any `json:"createdAt"` }
    out := make([]dto, 0, len(items))
    for _, it := range items {
        out = append(out, dto{ ID: it.ID.Hex(), Name: it.Name, DeviceKey: it.DeviceKey, Approved: it.Approved, ApprovedAt: it.ApprovedAt, CreatedAt: it.CreatedAt })
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{ "items": out })
}

// ListStationProducts godoc
// @Summary List station products
// @Tags admin-stations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Station ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/stations/{id}/products [get]
func (h *AdminStationHandler) ListStationProducts(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(idStr)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    if _, err := h.stations.FindByID(r.Context(), oid); err != nil { response.WriteError(w, http.StatusNotFound, "station not found"); return }
    pids, err := h.stationProds.ListProductIDsByStation(r.Context(), oid)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to list assignments"); return }
    prods, _ := h.products.GetByIDs(r.Context(), pids)
    type row struct{ ProductID string `json:"productId"`; Name string `json:"name"` }
    out := make([]row, 0, len(prods))
    for _, p := range prods { out = append(out, row{ ProductID: p.ID.Hex(), Name: p.Name }) }
    response.WriteJSON(w, http.StatusOK, map[string]any{"items": out})
}

// AssignProducts godoc
// @Summary Assign products to station
// @Tags admin-stations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Station ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/stations/{id}/products [post]
func (h *AdminStationHandler) AssignProducts(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(idStr)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    if _, err := h.stations.FindByID(r.Context(), oid); err != nil { response.WriteError(w, http.StatusNotFound, "station not found"); return }
    var body struct{ ProductIDs []string `json:"productIds"` }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.ProductIDs) == 0 { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    ids := make([]primitive.ObjectID, 0, len(body.ProductIDs))
    for _, s := range body.ProductIDs { if pid, e := primitive.ObjectIDFromHex(s); e == nil { ids = append(ids, pid) } }
    if len(ids) == 0 { response.WriteError(w, http.StatusBadRequest, "no valid product ids"); return }
    added, err := h.stationProds.AddProducts(r.Context(), oid, ids)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "assign failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "station", EntityID: idStr, Before: nil, After: map[string]any{"assigned": len(ids), "added": added} })
    response.WriteJSON(w, http.StatusOK, map[string]any{"assigned": added})
}

// RemoveProduct godoc
// @Summary Remove product from station
// @Tags admin-stations
// @Security BearerAuth
// @Param id path string true "Station ID"
// @Param productId path string true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/stations/{id}/products/{productId} [delete]
func (h *AdminStationHandler) RemoveProduct(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    pidStr := chi.URLParam(r, "productId")
    oid, err := primitive.ObjectIDFromHex(idStr)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    pid, err := primitive.ObjectIDFromHex(pidStr)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid product id"); return }
    if _, err := h.stations.FindByID(r.Context(), oid); err != nil { response.WriteError(w, http.StatusNotFound, "station not found"); return }
    ok, err := h.stationProds.RemoveProduct(r.Context(), oid, pid)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "remove failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "station", EntityID: idStr, Before: nil, After: map[string]any{"removedProductId": pidStr, "removed": ok} })
    response.WriteJSON(w, http.StatusOK, map[string]any{"removed": ok})
}

// Approve godoc
// @Summary Approve station request
// @Tags admin-stations
// @Security BearerAuth
// @Param id path string true "Request ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/stations/requests/{id}/approve [post]
func (h *AdminStationHandler) Approve(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    req, err := h.stationReqs.FindByID(r.Context(), oid)
    if err != nil { response.WriteError(w, http.StatusNotFound, "not found"); return }
    // mark station approved
    if err := h.stations.ApproveByDeviceKey(r.Context(), req.DeviceKey); err != nil { response.WriteError(w, http.StatusInternalServerError, "approve station failed"); return }
    // mark request approved
    var decidedBy *primitive.ObjectID
    if claims, ok := middleware.GetUserFromContext(r.Context()); ok && claims != nil && claims.Subject != "" { if u, err := primitive.ObjectIDFromHex(claims.Subject); err == nil { decidedBy = &u } }
    if err := h.stationReqs.UpdateStatus(r.Context(), oid, domain.StationRequestStatusApproved, decidedBy, nowUTC()); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    // audit best-effort
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "station_request", EntityID: id, Before: nil, After: map[string]any{"status":"approved"} })
    response.WriteJSON(w, http.StatusOK, response.Ack{ Message: "approved" })
}

// Reject godoc
// @Summary Reject station request
// @Tags admin-stations
// @Security BearerAuth
// @Param id path string true "Request ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/stations/requests/{id}/reject [post]
func (h *AdminStationHandler) Reject(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    if _, err := h.stationReqs.FindByID(r.Context(), oid); err != nil { response.WriteError(w, http.StatusNotFound, "not found"); return }
    var decidedBy *primitive.ObjectID
    if claims, ok := middleware.GetUserFromContext(r.Context()); ok && claims != nil && claims.Subject != "" { if u, err := primitive.ObjectIDFromHex(claims.Subject); err == nil { decidedBy = &u } }
    if err := h.stationReqs.UpdateStatus(r.Context(), oid, domain.StationRequestStatusRejected, decidedBy, nowUTC()); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "station_request", EntityID: id, Before: nil, After: map[string]any{"status":"rejected"} })
    response.WriteJSON(w, http.StatusOK, response.Ack{ Message: "rejected" })
}

func nowUTC() time.Time { return time.Now().UTC() }
