package handler

import (
    "backend/internal/domain"
    "backend/internal/middleware"
    "backend/internal/repository"
    "backend/internal/response"
    "encoding/json"
    "net/http"

    "github.com/go-playground/validator/v10"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminProductHandler struct {
    products repository.ProductRepository
    inventory repository.InventoryLedgerRepository
    audit repository.AuditRepository
    validator *validator.Validate
    slotItems repository.MenuSlotItemRepository
    categories repository.CategoryRepository
}

func NewAdminProductHandler(prod repository.ProductRepository, inv repository.InventoryLedgerRepository, audit repository.AuditRepository, slotItems repository.MenuSlotItemRepository, cats repository.CategoryRepository) *AdminProductHandler {
    return &AdminProductHandler{ products: prod, inventory: inv, audit: audit, validator: validator.New(), slotItems: slotItems, categories: cats }
}

type patchPriceBody struct{ PriceCents int64 `json:"priceCents" validate:"required,gte=0"` }

func (h *AdminProductHandler) PatchPrice(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserFromContext(r.Context())
    if !ok { response.WriteError(w, http.StatusUnauthorized, "unauthorized"); return }
    id := chiURLParam(r, "id")
    pid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    before, _ := h.products.FindByID(r.Context(), pid)
    if before == nil { response.WriteError(w, http.StatusNotFound, "not found"); return }
    var body patchPriceBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid json"); return }
    if err := h.validator.Struct(body); err != nil { response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path)); return }
    if err := h.products.UpdateFields(r.Context(), pid, primitive.M{"price_cents": body.PriceCents}); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    after, _ := h.products.FindByID(r.Context(), pid)
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{
        Action:     domain.AuditUpdate,
        EntityType: "product",
        EntityID:   id,
        Before:     before,
        After:      after,
        RequestID:  getRequestIDPtr(r),
        ActorUserID: objIDPtr(claims.Subject),
        ActorRole:  strPtr(string(claims.Role)),
    })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type patchActiveBody struct{ IsActive bool `json:"isActive"` }

func (h *AdminProductHandler) PatchActive(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserFromContext(r.Context())
    if !ok { response.WriteError(w, http.StatusUnauthorized, "unauthorized"); return }
    id := chiURLParam(r, "id")
    pid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    before, _ := h.products.FindByID(r.Context(), pid)
    if before == nil { response.WriteError(w, http.StatusNotFound, "not found"); return }
    var body patchActiveBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid json"); return }
    if err := h.products.UpdateFields(r.Context(), pid, primitive.M{"is_active": body.IsActive}); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    after, _ := h.products.FindByID(r.Context(), pid)
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "product", EntityID: id, Before: before, After: after, RequestID: getRequestIDPtr(r), ActorUserID: objIDPtr(claims.Subject), ActorRole: strPtr(string(claims.Role)) })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type inventoryAdjustBody struct{ Delta int `json:"delta" validate:"required,ne=0"`; Reason domain.InventoryReason `json:"reason" validate:"required,oneof=opening_balance sale refund manual_adjust correction"` }

func (h *AdminProductHandler) AdjustInventory(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserFromContext(r.Context())
    if !ok { response.WriteError(w, http.StatusUnauthorized, "unauthorized"); return }
    id := chiURLParam(r, "id")
    pid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    before, _ := h.products.FindByID(r.Context(), pid) // snapshot only
    var body inventoryAdjustBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid json"); return }
    if err := h.validator.Struct(body); err != nil { response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path)); return }
    if h.inventory == nil { response.WriteError(w, http.StatusNotImplemented, "inventory disabled"); return }
    if err := h.inventory.Append(r.Context(), &domain.InventoryLedger{ ProductID: pid, Delta: body.Delta, Reason: body.Reason }); err != nil {
        response.WriteError(w, http.StatusInternalServerError, "adjust failed"); return
    }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "inventory", EntityID: id, Before: before, After: map[string]any{"delta": body.Delta, "reason": body.Reason}, RequestID: getRequestIDPtr(r), ActorUserID: objIDPtr(claims.Subject), ActorRole: strPtr(string(claims.Role)) })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DELETE /v1/admin/products/{id} - hard delete; only allowed if product is not referenced in menu compositions
func (h *AdminProductHandler) DeleteHard(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id")
    pid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    p, err := h.products.FindByID(r.Context(), pid)
    if err != nil || p == nil { response.WriteError(w, http.StatusNotFound, "not found"); return }
    // Prevent deleting if attached to any menu slot
    if h.slotItems != nil {
        if n, err := h.slotItems.CountByProductID(r.Context(), pid); err == nil && n > 0 { response.WriteError(w, http.StatusConflict, "product is used in menu compositions"); return }
    }
    if err := h.products.DeleteByID(r.Context(), pid); err != nil { response.WriteError(w, http.StatusInternalServerError, "delete failed"); return }
    response.WriteNoContent(w)
}

// PATCH /v1/admin/products/{id}/category - move product between categories
type patchCategoryBody struct { CategoryID string `json:"categoryId"` }
func (h *AdminProductHandler) PatchCategory(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); pid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    var body patchCategoryBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.CategoryID == "" { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    cid, err := primitive.ObjectIDFromHex(body.CategoryID); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid category"); return }
    if h.categories != nil {
        c, err := h.categories.GetByID(r.Context(), cid); if err != nil || c == nil { response.WriteError(w, http.StatusBadRequest, "category not found"); return }
    }
    before, _ := h.products.FindByID(r.Context(), pid)
    if err := h.products.UpdateFields(r.Context(), pid, primitive.M{"category_id": cid}); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    after, _ := h.products.FindByID(r.Context(), pid)
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "product", EntityID: id, Before: before, After: after, RequestID: getRequestIDPtr(r) })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// helpers
func chiURLParam(r *http.Request, name string) string { return ChiURLParamFn(r, name) }

// indirection to avoid pulling chi here; router will assign ChiURLParamFn at init
var ChiURLParamFn = func(r *http.Request, name string) string { return "" }

func strPtr(s string) *string { if s == "" { return nil }; return &s }
func objIDPtr(hex string) *primitive.ObjectID {
    id, err := primitive.ObjectIDFromHex(hex); if err != nil { return nil }
    return &id
}

func getRequestIDPtr(r *http.Request) *string {
    id := r.Header.Get("X-Request-ID")
    if id == "" { return nil }
    return &id
}
