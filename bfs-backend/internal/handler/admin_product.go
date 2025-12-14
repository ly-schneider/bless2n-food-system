package handler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/service"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AdminProductHandler struct {
	products   repository.ProductRepository
	inventory  repository.InventoryLedgerRepository
	audit      repository.AuditRepository
	validator  *validator.Validate
	slotItems  repository.MenuSlotItemRepository
	categories repository.CategoryRepository
	posConfig  service.POSConfigService
}

func NewAdminProductHandler(prod repository.ProductRepository, inv repository.InventoryLedgerRepository, audit repository.AuditRepository, slotItems repository.MenuSlotItemRepository, cats repository.CategoryRepository, posConfig service.POSConfigService) *AdminProductHandler {
	return &AdminProductHandler{products: prod, inventory: inv, audit: audit, validator: validator.New(), slotItems: slotItems, categories: cats, posConfig: posConfig}
}

type createProductBody struct {
	Name       string             `json:"name" validate:"required,min=1"`
	PriceCents int64              `json:"priceCents" validate:"required,gte=0"`
	CategoryID string             `json:"categoryId" validate:"required"`
	Type       domain.ProductType `json:"type" validate:"required,oneof=simple menu"`
	Image      *string            `json:"image,omitempty"`
	IsActive   *bool              `json:"isActive,omitempty"`
	JetonID    *string            `json:"jetonId,omitempty"`
}

// Create godoc
// @Summary Create product
// @Tags admin-products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body createProductBody true "Product payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Failure 401 {object} response.ProblemDetails
// @Router /v1/admin/products [post]
func (h *AdminProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body createProductBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := h.validator.Struct(body); err != nil {
		response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path))
		return
	}
	if body.Type != domain.ProductTypeSimple {
		response.WriteError(w, http.StatusBadRequest, "only simple products can be created here")
		return
	}
	catID, err := bson.ObjectIDFromHex(body.CategoryID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid category")
		return
	}
	if h.categories != nil {
		c, err := h.categories.GetByID(r.Context(), catID)
		if err != nil || c == nil {
			response.WriteError(w, http.StatusBadRequest, "category not found")
			return
		}
	}
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	var jetonOID *bson.ObjectID
	if body.JetonID != nil && *body.JetonID != "" {
		if oid, err := bson.ObjectIDFromHex(*body.JetonID); err == nil {
			jetonOID = &oid
		} else {
			response.WriteError(w, http.StatusBadRequest, "invalid jeton id")
			return
		}
	}
	if active && h.posConfig != nil {
		if settings, err := h.posConfig.GetSettings(r.Context()); err == nil && settings != nil && settings.Mode == domain.PosModeJeton {
			if jetonOID == nil {
				response.WriteError(w, http.StatusBadRequest, "jeton_required")
				return
			}
		}
	}
	p := &domain.Product{
		CategoryID: catID,
		Type:       domain.ProductTypeSimple,
		Name:       body.Name,
		Image:      body.Image,
		PriceCents: domain.Cents(body.PriceCents),
		JetonID:    jetonOID,
		IsActive:   active,
	}
	id, err := h.products.Insert(r.Context(), p)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "create failed")
		return
	}
	_ = h.audit.Insert(r.Context(), &domain.AuditLog{
		Action:      domain.AuditCreate,
		EntityType:  "product",
		EntityID:    id.Hex(),
		After:       p,
		RequestID:   getRequestIDPtr(r),
		ActorUserID: objIDPtr(claims.Subject),
		ActorRole:   strPtr(string(claims.Role)),
	})
	response.WriteJSON(w, http.StatusCreated, map[string]any{"id": id.Hex()})
}

type patchPriceBody struct {
	PriceCents int64 `json:"priceCents" validate:"required,gte=0"`
}

// PatchPrice godoc
// @Summary Update product price
// @Tags admin-products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param payload body patchPriceBody true "Price payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Failure 401 {object} response.ProblemDetails
// @Failure 404 {object} response.ProblemDetails
// @Router /v1/admin/products/{id}/price [patch]
func (h *AdminProductHandler) PatchPrice(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chiURLParam(r, "id")
	pid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	before, _ := h.products.FindByID(r.Context(), pid)
	if before == nil {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	var body patchPriceBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.validator.Struct(body); err != nil {
		response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path))
		return
	}
	if err := h.products.UpdateFields(r.Context(), pid, bson.M{"price_cents": body.PriceCents}); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	after, _ := h.products.FindByID(r.Context(), pid)
	_ = h.audit.Insert(r.Context(), &domain.AuditLog{
		Action:      domain.AuditUpdate,
		EntityType:  "product",
		EntityID:    id,
		Before:      before,
		After:       after,
		RequestID:   getRequestIDPtr(r),
		ActorUserID: objIDPtr(claims.Subject),
		ActorRole:   strPtr(string(claims.Role)),
	})
	response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type patchActiveBody struct {
	IsActive bool `json:"isActive"`
}

// PatchActive godoc
// @Summary Set product active flag
// @Tags admin-products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param payload body patchActiveBody true "Active payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Failure 401 {object} response.ProblemDetails
// @Failure 404 {object} response.ProblemDetails
// @Router /v1/admin/products/{id}/active [patch]
func (h *AdminProductHandler) PatchActive(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chiURLParam(r, "id")
	pid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	before, _ := h.products.FindByID(r.Context(), pid)
	if before == nil {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	var body patchActiveBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.IsActive && h.posConfig != nil {
		if settings, err := h.posConfig.GetSettings(r.Context()); err == nil && settings != nil && settings.Mode == domain.PosModeJeton {
			if before.JetonID == nil || before.JetonID.IsZero() {
				response.WriteError(w, http.StatusBadRequest, "jeton_required")
				return
			}
		}
	}
	if err := h.products.UpdateFields(r.Context(), pid, bson.M{"is_active": body.IsActive}); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	after, _ := h.products.FindByID(r.Context(), pid)
	_ = h.audit.Insert(r.Context(), &domain.AuditLog{Action: domain.AuditUpdate, EntityType: "product", EntityID: id, Before: before, After: after, RequestID: getRequestIDPtr(r), ActorUserID: objIDPtr(claims.Subject), ActorRole: strPtr(string(claims.Role))})
	response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type patchJetonBody struct {
	JetonID *string `json:"jetonId"`
}

// PatchJeton godoc
// @Summary Assign jeton to product
// @Tags admin-products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param payload body patchJetonBody true "Jeton payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/admin/products/{id}/jeton [patch]
func (h *AdminProductHandler) PatchJeton(w http.ResponseWriter, r *http.Request) {
	if h.posConfig == nil {
		response.WriteError(w, http.StatusInternalServerError, "jeton service unavailable")
		return
	}
	id := chiURLParam(r, "id")
	pid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body patchJetonBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var jetonOID *bson.ObjectID
	if body.JetonID != nil && *body.JetonID != "" {
		if oid, err := bson.ObjectIDFromHex(*body.JetonID); err == nil {
			jetonOID = &oid
		} else {
			response.WriteError(w, http.StatusBadRequest, "invalid jeton id")
			return
		}
	}
	if err := h.posConfig.SetProductJeton(r.Context(), pid, jetonOID); err != nil {
		if errors.Is(err, service.ErrJetonRequired) {
			response.WriteError(w, http.StatusBadRequest, "jeton_required")
			return
		}
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type inventoryAdjustBody struct {
	Delta  int                    `json:"delta" validate:"required,ne=0"`
	Reason domain.InventoryReason `json:"reason" validate:"required,oneof=opening_balance sale refund manual_adjust correction"`
}

// AdjustInventory godoc
// @Summary Adjust product inventory
// @Tags admin-products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param payload body inventoryAdjustBody true "Adjustment payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Failure 401 {object} response.ProblemDetails
// @Router /v1/admin/products/{id}/inventory-adjust [post]
func (h *AdminProductHandler) AdjustInventory(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chiURLParam(r, "id")
	pid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	before, _ := h.products.FindByID(r.Context(), pid) // snapshot only
	var body inventoryAdjustBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.validator.Struct(body); err != nil {
		response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path))
		return
	}
	if h.inventory == nil {
		response.WriteError(w, http.StatusNotImplemented, "inventory disabled")
		return
	}
	if err := h.inventory.Append(r.Context(), &domain.InventoryLedger{ProductID: pid, Delta: body.Delta, Reason: body.Reason}); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "adjust failed")
		return
	}
	_ = h.audit.Insert(r.Context(), &domain.AuditLog{Action: domain.AuditUpdate, EntityType: "inventory", EntityID: id, Before: before, After: map[string]any{"delta": body.Delta, "reason": body.Reason}, RequestID: getRequestIDPtr(r), ActorUserID: objIDPtr(claims.Subject), ActorRole: strPtr(string(claims.Role))})
	response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DeleteHard godoc
// @Summary Hard delete product
// @Description Only allowed if product is not referenced in menu compositions
// @Tags admin-products
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ProblemDetails
// @Failure 404 {object} response.ProblemDetails
// @Router /v1/admin/products/{id} [delete]
func (h *AdminProductHandler) DeleteHard(w http.ResponseWriter, r *http.Request) {
	id := chiURLParam(r, "id")
	pid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := h.products.FindByID(r.Context(), pid)
	if err != nil || p == nil {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	// Prevent deleting if attached to any menu slot
	if h.slotItems != nil {
		if n, err := h.slotItems.CountByProductID(r.Context(), pid); err == nil && n > 0 {
			response.WriteError(w, http.StatusConflict, "product is used in menu compositions")
			return
		}
	}
	if err := h.products.DeleteByID(r.Context(), pid); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	response.WriteNoContent(w)
}

// PatchCategory godoc
// @Summary Move product to category
// @Tags admin-products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param payload body patchCategoryBody true "Category payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Failure 404 {object} response.ProblemDetails
// @Router /v1/admin/products/{id}/category [patch]
type patchCategoryBody struct {
	CategoryID string `json:"categoryId"`
}

func (h *AdminProductHandler) PatchCategory(w http.ResponseWriter, r *http.Request) {
	id := chiURLParam(r, "id")
	pid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body patchCategoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.CategoryID == "" {
		response.WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	cid, err := bson.ObjectIDFromHex(body.CategoryID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid category")
		return
	}
	if h.categories != nil {
		c, err := h.categories.GetByID(r.Context(), cid)
		if err != nil || c == nil {
			response.WriteError(w, http.StatusBadRequest, "category not found")
			return
		}
	}
	before, _ := h.products.FindByID(r.Context(), pid)
	if err := h.products.UpdateFields(r.Context(), pid, bson.M{"category_id": cid}); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	after, _ := h.products.FindByID(r.Context(), pid)
	_ = h.audit.Insert(r.Context(), &domain.AuditLog{Action: domain.AuditUpdate, EntityType: "product", EntityID: id, Before: before, After: after, RequestID: getRequestIDPtr(r)})
	response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// helpers
func chiURLParam(r *http.Request, name string) string { return ChiURLParamFn(r, name) }

// indirection to avoid pulling chi here; router will assign ChiURLParamFn at init
var ChiURLParamFn = func(r *http.Request, name string) string { return "" }

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
func objIDPtr(hex string) *bson.ObjectID {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		return nil
	}
	return &id
}

func getRequestIDPtr(r *http.Request) *string {
	id := r.Header.Get("X-Request-ID")
	if id == "" {
		return nil
	}
	return &id
}
