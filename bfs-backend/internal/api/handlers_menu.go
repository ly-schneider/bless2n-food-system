package api

import (
	"encoding/json"
	"net/http"

	"backend/internal/generated/api/generated"
	"backend/internal/generated/ent"
	"backend/internal/generated/ent/product"
	"backend/internal/response"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ListMenus returns all menu-type products with their slots and options.
// (GET /menus)
func (h *Handlers) ListMenus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	menus, err := h.products.GetMenus(ctx)
	if err != nil {
		writeEntError(w, err)
		return
	}

	items := make([]generated.Menu, 0, len(menus))
	for _, m := range menus {
		items = append(items, entProductToAPIMenu(m))
	}
	response.WriteJSON(w, http.StatusOK, generated.MenuList{Items: items})
}

// CreateMenu creates a new menu (product with type=menu).
// (POST /menus)
func (h *Handlers) CreateMenu(w http.ResponseWriter, r *http.Request) {
	var body generated.MenuCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	var priceCents int64
	if body.PriceCents != nil {
		priceCents = *body.PriceCents
	}

	prod, err := h.products.Create(
		r.Context(),
		uuid.UUID(body.CategoryId),
		product.TypeMenu,
		body.Name,
		priceCents,
		true, // isActive
		body.Image,
		nil, // no jeton for menus
	)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, entProductToAPIMenu(prod))
}

// GetMenu returns a single menu by ID with slots and options loaded.
// (GET /menus/{menuId})
func (h *Handlers) GetMenu(w http.ResponseWriter, r *http.Request, menuId openapi_types.UUID) {
	prod, err := h.products.GetByID(r.Context(), uuid.UUID(menuId))
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, entProductToAPIMenu(prod))
}

// UpdateMenu partially updates a menu product.
// (PATCH /menus/{menuId})
func (h *Handlers) UpdateMenu(w http.ResponseWriter, r *http.Request, menuId openapi_types.UUID) {
	ctx := r.Context()
	id := uuid.UUID(menuId)

	existing, err := h.products.GetByID(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}

	var body generated.MenuUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	name := existing.Name
	if body.Name != nil {
		name = *body.Name
	}
	priceCents := existing.PriceCents
	if body.PriceCents != nil {
		priceCents = *body.PriceCents
	}
	isActive := existing.IsActive
	if body.IsActive != nil {
		isActive = *body.IsActive
	}
	image := existing.Image
	if body.Image != nil {
		image = body.Image
	}

	prod, err := h.products.Update(
		ctx,
		id,
		existing.CategoryID,
		product.TypeMenu,
		name,
		priceCents,
		isActive,
		image,
		nil, // no jeton for menus
	)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, entProductToAPIMenu(prod))
}

// DeleteMenu removes a menu product.
// (DELETE /menus/{menuId})
func (h *Handlers) DeleteMenu(w http.ResponseWriter, r *http.Request, menuId openapi_types.UUID) {
	if err := h.products.Delete(r.Context(), uuid.UUID(menuId)); err != nil {
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateMenuSlot creates a new slot on a menu.
// (POST /menus/{menuId}/slots)
func (h *Handlers) CreateMenuSlot(w http.ResponseWriter, r *http.Request, menuId openapi_types.UUID) {
	var body generated.MenuSlotCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	slot, err := h.products.CreateMenuSlot(r.Context(), uuid.UUID(menuId), body.Name)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, toAPIMenuSlot(slot))
}

// UpdateMenuSlot renames a menu slot.
// (PATCH /menus/{menuId}/slots/{slotId})
func (h *Handlers) UpdateMenuSlot(w http.ResponseWriter, r *http.Request, menuId openapi_types.UUID, slotId openapi_types.UUID) {
	var body generated.MenuSlotUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	name := ""
	if body.Name != nil {
		name = *body.Name
	}

	slot, err := h.products.UpdateMenuSlot(r.Context(), uuid.UUID(menuId), uuid.UUID(slotId), name)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIMenuSlot(slot))
}

// DeleteMenuSlot removes a slot from a menu.
// (DELETE /menus/{menuId}/slots/{slotId})
func (h *Handlers) DeleteMenuSlot(w http.ResponseWriter, r *http.Request, menuId openapi_types.UUID, slotId openapi_types.UUID) {
	if err := h.products.DeleteMenuSlot(r.Context(), uuid.UUID(menuId), uuid.UUID(slotId)); err != nil {
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ReorderMenuSlots updates the sequence order of menu slots.
// (PATCH /menus/{menuId}/slots/reorder)
func (h *Handlers) ReorderMenuSlots(w http.ResponseWriter, r *http.Request, menuId openapi_types.UUID) {
	var body generated.MenuSlotReorder
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	// Convert string keys to uuid.UUID keys.
	positions := make(map[uuid.UUID]int, len(body.Positions))
	for idStr, seq := range body.Positions {
		uid, err := uuid.Parse(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_slot_id", "Invalid slot UUID: "+idStr)
			return
		}
		positions[uid] = seq
	}

	if err := h.products.ReorderMenuSlots(r.Context(), uuid.UUID(menuId), positions); err != nil {
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddSlotOption adds a product as an option to a menu slot.
// (POST /menus/{menuId}/slots/{slotId}/options)
func (h *Handlers) AddSlotOption(w http.ResponseWriter, r *http.Request, menuId openapi_types.UUID, slotId openapi_types.UUID) {
	var body generated.MenuSlotOptionCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	opt, err := h.products.AddSlotOption(r.Context(), uuid.UUID(menuId), uuid.UUID(slotId), uuid.UUID(body.ProductId))
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, toAPIMenuSlotOption(opt))
}

// RemoveSlotOption removes a product option from a menu slot.
// (DELETE /menus/{menuId}/slots/{slotId}/options/{optionProductId})
func (h *Handlers) RemoveSlotOption(w http.ResponseWriter, r *http.Request, menuId openapi_types.UUID, slotId openapi_types.UUID, optionProductId openapi_types.UUID) {
	if err := h.products.RemoveSlotOption(r.Context(), uuid.UUID(menuId), uuid.UUID(slotId), uuid.UUID(optionProductId)); err != nil {
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// entProductToAPIMenu converts an ent.Product (type=menu) into a generated.Menu.
func entProductToAPIMenu(e *ent.Product) generated.Menu {
	m := generated.Menu{
		Id:         e.ID,
		CategoryId: e.CategoryID,
		Name:       e.Name,
		Image:      e.Image,
		PriceCents: e.PriceCents,
		IsActive:   e.IsActive,
		CreatedAt:  ptr(e.CreatedAt),
		UpdatedAt:  ptr(e.UpdatedAt),
	}

	// Map MenuSlots edge if loaded.
	if slots, err := e.Edges.MenuSlotsOrErr(); err == nil && len(slots) > 0 {
		apiSlots := make([]generated.MenuSlot, 0, len(slots))
		for _, s := range slots {
			apiSlots = append(apiSlots, toAPIMenuSlot(s))
		}
		m.Slots = &apiSlots
	}

	return m
}
