package handler

import (
    "backend/internal/domain"
    "backend/internal/middleware"
    "backend/internal/repository"
    "backend/internal/response"
    "encoding/json"
    "net/http"
    "sort"
    "strconv"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminMenuHandler struct {
    products repository.ProductRepository
    categories repository.CategoryRepository
    slots repository.MenuSlotRepository
    items repository.MenuSlotItemRepository
    audit repository.AuditRepository
}

func NewAdminMenuHandler(prod repository.ProductRepository, cat repository.CategoryRepository, slots repository.MenuSlotRepository, items repository.MenuSlotItemRepository, audit repository.AuditRepository) *AdminMenuHandler {
    return &AdminMenuHandler{ products: prod, categories: cat, slots: slots, items: items, audit: audit }
}

// List menus (products of type menu)
func (h *AdminMenuHandler) List(w http.ResponseWriter, r *http.Request) {
    var active *bool
    if v := r.URL.Query().Get("active"); v != "" {
        b := v == "true" || v == "1"
        active = &b
    }
    var q *string
    if v := r.URL.Query().Get("q"); v != "" { q = &v }
    limit := 50; offset := 0
    if v := r.URL.Query().Get("limit"); v != "" { if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 { limit = n } }
    if v := r.URL.Query().Get("offset"); v != "" { if n, err := strconv.Atoi(v); err == nil && n >= 0 { offset = n } }
    menus, total, err := h.products.GetMenus(r.Context(), q, active, limit, offset)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to list menus"); return }
    type MenuDTO struct { ID string `json:"id"`; Name string `json:"name"`; PriceCents int64 `json:"priceCents"`; IsActive bool `json:"isActive"`; Image *string `json:"image,omitempty"`}
    out := make([]MenuDTO, 0, len(menus))
    for _, m := range menus { out = append(out, MenuDTO{ ID: m.ID.Hex(), Name: m.Name, PriceCents: int64(m.PriceCents), IsActive: m.IsActive, Image: m.Image }) }
    response.WriteJSON(w, http.StatusOK, map[string]any{"items": out, "count": total})
}

// Get menu detail with slots and items
func (h *AdminMenuHandler) Get(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    m, err := h.products.FindByID(r.Context(), oid)
    if err != nil || m == nil || m.Type != domain.ProductTypeMenu { response.WriteError(w, http.StatusNotFound, "not found"); return }
    slots, err := h.slots.FindByProductIDs(r.Context(), []primitive.ObjectID{m.ID})
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to load slots"); return }
    sort.Slice(slots, func(i,j int) bool { return slots[i].Sequence < slots[j].Sequence })
    slotIDs := make([]primitive.ObjectID, 0, len(slots))
    for _, s := range slots { slotIDs = append(slotIDs, s.ID) }
    items, err := h.items.FindByMenuSlotIDs(r.Context(), slotIDs)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to load items"); return }
    pidset := map[primitive.ObjectID]struct{}{}
    for _, it := range items { pidset[it.ProductID] = struct{}{} }
    pids := make([]primitive.ObjectID, 0, len(pidset))
    for pid := range pidset { pids = append(pids, pid) }
    prods, _ := h.products.GetByIDs(r.Context(), pids)
    prodByID := map[primitive.ObjectID]*domain.Product{}
    for _, p := range prods { prodByID[p.ID] = p }
    // Build DTO
    type SlotItemDTO = domain.ProductSummaryDTO
    type SlotDTO struct { ID string `json:"id"`; Name string `json:"name"`; Sequence int `json:"sequence"`; MenuSlotItems []SlotItemDTO `json:"menuSlotItems"` }
    outSlots := make([]SlotDTO, 0, len(slots))
    for _, s := range slots {
        var sitems []SlotItemDTO
        for _, it := range items {
            if it.MenuSlotID == s.ID {
                if p := prodByID[it.ProductID]; p != nil {
                    sitems = append(sitems, domain.ProductSummaryDTO{ ID: p.ID.Hex(), Category: domain.CategoryDTO{ ID: p.CategoryID.Hex(), Name: "", IsActive: false }, Type: domain.ProductType(p.Type), Name: p.Name, Image: p.Image, PriceCents: domain.Cents(p.PriceCents), IsActive: p.IsActive })
                }
            }
        }
        outSlots = append(outSlots, SlotDTO{ ID: s.ID.Hex(), Name: s.Name, Sequence: s.Sequence, MenuSlotItems: sitems })
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{"id": m.ID.Hex(), "name": m.Name, "priceCents": int64(m.PriceCents), "isActive": m.IsActive, "image": m.Image, "slots": outSlots})
}

type createMenuBody struct { Name string `json:"name"`; PriceCents int64 `json:"priceCents"`; CategoryID string `json:"categoryId"`; Image *string `json:"image,omitempty"`; IsActive *bool `json:"isActive,omitempty"` }

func (h *AdminMenuHandler) Create(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserFromContext(r.Context()); if !ok { response.WriteError(w, http.StatusUnauthorized, "unauthorized"); return }
    var body createMenuBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.PriceCents < 0 || body.CategoryID == "" { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    catID, err := primitive.ObjectIDFromHex(body.CategoryID); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid category"); return }
    cat, err := h.categories.GetByID(r.Context(), catID); if err != nil || cat == nil { response.WriteError(w, http.StatusBadRequest, "category not found"); return }
    p := &domain.Product{ CategoryID: cat.ID, Type: domain.ProductTypeMenu, Name: body.Name, Image: body.Image, PriceCents: domain.Cents(body.PriceCents), IsActive: true }
    if body.IsActive != nil { p.IsActive = *body.IsActive }
    id, err := h.products.Insert(r.Context(), p); if err != nil { response.WriteError(w, http.StatusInternalServerError, "create failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditCreate, EntityType: "menu", EntityID: id.Hex(), After: p, ActorUserID: objIDPtr(claims.Subject), ActorRole: strPtr(string(claims.Role)), RequestID: getRequestIDPtr(r) })
    response.WriteJSON(w, http.StatusCreated, map[string]any{"id": id.Hex()})
}

type updateMenuBody struct { Name *string `json:"name,omitempty"`; PriceCents *int64 `json:"priceCents,omitempty"`; Image *string `json:"image,omitempty"`; IsActive *bool `json:"isActive,omitempty"` }

func (h *AdminMenuHandler) Update(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    var body updateMenuBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    set := primitive.M{}
    if body.Name != nil { set["name"] = *body.Name }
    if body.PriceCents != nil && *body.PriceCents >= 0 { set["price_cents"] = *body.PriceCents }
    if body.Image != nil { if *body.Image == "" { set["image"] = nil } else { set["image"] = *body.Image } }
    if body.IsActive != nil { set["is_active"] = *body.IsActive }
    if len(set) == 0 { response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true}); return }
    before, _ := h.products.FindByID(r.Context(), oid)
    if err := h.products.UpdateFields(r.Context(), oid, set); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    after, _ := h.products.FindByID(r.Context(), oid)
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "menu", EntityID: id, Before: before, After: after, RequestID: getRequestIDPtr(r) })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Soft delete: set inactive
func (h *AdminMenuHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    if err := h.products.UpdateFields(r.Context(), oid, primitive.M{"is_active": false}); err != nil { response.WriteError(w, http.StatusInternalServerError, "delete failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "menu", EntityID: id, Before: map[string]any{"id": id}, After: map[string]any{"is_active": false}, RequestID: getRequestIDPtr(r) })
    response.WriteNoContent(w)
}

// PATCH /v1/admin/menus/{id}/active - toggle is_active
type patchMenuActiveBody struct { IsActive bool `json:"isActive"` }
func (h *AdminMenuHandler) PatchActive(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    var body patchMenuActiveBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    before, _ := h.products.FindByID(r.Context(), oid)
    if err := h.products.UpdateFields(r.Context(), oid, primitive.M{"is_active": body.IsActive}); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    after, _ := h.products.FindByID(r.Context(), oid)
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "menu", EntityID: id, Before: before, After: after, RequestID: getRequestIDPtr(r) })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DELETE /v1/admin/menus/{id}/hard - permanently delete the menu product and its slots/items
func (h *AdminMenuHandler) DeleteHard(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    // delete slot items then slots then product
    slots, err := h.slots.FindByProductIDs(r.Context(), []primitive.ObjectID{oid}); if err != nil { response.WriteError(w, http.StatusInternalServerError, "load slots failed"); return }
    for _, s := range slots {
        _ = h.items.DeleteByMenuSlotID(r.Context(), s.ID)
        _ = h.slots.DeleteByID(r.Context(), s.ID)
    }
    if err := h.products.DeleteByID(r.Context(), oid); err != nil { response.WriteError(w, http.StatusInternalServerError, "delete failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditDelete, EntityType: "menu", EntityID: id, RequestID: getRequestIDPtr(r) })
    response.WriteNoContent(w)
}

type createSlotBody struct { Name string `json:"name"`; Sequence *int `json:"sequence,omitempty"` }

func (h *AdminMenuHandler) CreateSlot(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    var body createSlotBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    seq := 1
    if body.Sequence != nil { seq = *body.Sequence }
    sid, err := h.slots.Insert(r.Context(), &domain.MenuSlot{ ProductID: oid, Name: body.Name, Sequence: seq })
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "create slot failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditCreate, EntityType: "menu_slot", EntityID: sid.Hex(), After: map[string]any{"name": body.Name, "sequence": seq, "menuId": id}, RequestID: getRequestIDPtr(r) })
    response.WriteJSON(w, http.StatusCreated, map[string]any{"id": sid.Hex()})
}

type renameSlotBody struct { Name string `json:"name"` }

func (h *AdminMenuHandler) RenameSlot(w http.ResponseWriter, r *http.Request) {
    sid := chiURLParam(r, "slotId"); soid, err := primitive.ObjectIDFromHex(sid); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    var body renameSlotBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    if err := h.slots.UpdateName(r.Context(), soid, body.Name); err != nil { response.WriteError(w, http.StatusInternalServerError, "rename failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "menu_slot", EntityID: sid, After: map[string]any{"name": body.Name}, RequestID: getRequestIDPtr(r) })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type reorderSlotsBody struct { Order []struct{ SlotID string `json:"slotId"`; Sequence int `json:"sequence"` } `json:"order"` }

func (h *AdminMenuHandler) ReorderSlots(w http.ResponseWriter, r *http.Request) {
    var body reorderSlotsBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Order) == 0 { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    seqs := map[primitive.ObjectID]int{}
    for _, it := range body.Order { if oid, err := primitive.ObjectIDFromHex(it.SlotID); err == nil { seqs[oid] = it.Sequence } }
    if err := h.slots.UpdateSequences(r.Context(), seqs); err != nil { response.WriteError(w, http.StatusInternalServerError, "reorder failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "menu_slots", EntityID: chiURLParam(r, "id"), After: body, RequestID: getRequestIDPtr(r) })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *AdminMenuHandler) DeleteSlot(w http.ResponseWriter, r *http.Request) {
    sid := chiURLParam(r, "slotId"); soid, err := primitive.ObjectIDFromHex(sid); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    if err := h.items.DeleteByMenuSlotID(r.Context(), soid); err != nil { response.WriteError(w, http.StatusInternalServerError, "delete items failed"); return }
    if err := h.slots.DeleteByID(r.Context(), soid); err != nil { response.WriteError(w, http.StatusInternalServerError, "delete slot failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditDelete, EntityType: "menu_slot", EntityID: sid, RequestID: getRequestIDPtr(r) })
    response.WriteNoContent(w)
}

type attachItemBody struct { ProductID string `json:"productId"` }

func (h *AdminMenuHandler) AttachItem(w http.ResponseWriter, r *http.Request) {
    sid := chiURLParam(r, "slotId"); soid, err := primitive.ObjectIDFromHex(sid); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid slot"); return }
    var body attachItemBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProductID == "" { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    pid, err := primitive.ObjectIDFromHex(body.ProductID); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid product"); return }
    // Validate product type
    p, err := h.products.FindByID(r.Context(), pid); if err != nil || p == nil { response.WriteError(w, http.StatusBadRequest, "product not found"); return }
    if p.Type != domain.ProductTypeSimple { response.WriteError(w, http.StatusBadRequest, "only simple products can be attached"); return }
    exists, err := h.items.ExistsBySlotAndProduct(r.Context(), soid, pid); if err != nil { response.WriteError(w, http.StatusInternalServerError, "check failed"); return }
    if exists { response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true}) ; return }
    if err := h.items.Insert(r.Context(), &domain.MenuSlotItem{ MenuSlotID: soid, ProductID: pid }); err != nil { response.WriteError(w, http.StatusInternalServerError, "attach failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditCreate, EntityType: "menu_slot_item", EntityID: soid.Hex()+":"+pid.Hex(), After: map[string]any{"slotId": soid.Hex(), "productId": pid.Hex()}, RequestID: getRequestIDPtr(r) })
    response.WriteJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

func (h *AdminMenuHandler) DetachItem(w http.ResponseWriter, r *http.Request) {
    sid := chiURLParam(r, "slotId"); soid, err := primitive.ObjectIDFromHex(sid); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid slot"); return }
    pidStr := chiURLParam(r, "productId"); pid, err := primitive.ObjectIDFromHex(pidStr); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid product"); return }
    _, err = h.items.DeleteBySlotAndProduct(r.Context(), soid, pid); if err != nil { response.WriteError(w, http.StatusInternalServerError, "detach failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditDelete, EntityType: "menu_slot_item", EntityID: soid.Hex()+":"+pid.Hex(), RequestID: getRequestIDPtr(r) })
    response.WriteNoContent(w)
}
