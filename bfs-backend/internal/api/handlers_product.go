package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"backend/internal/generated/api/generated"
	"backend/internal/generated/ent"
	"backend/internal/generated/ent/product"
	"backend/internal/response"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

// ListProducts returns products, optionally filtered by category.
// (GET /products)
func (h *Handlers) ListProducts(w http.ResponseWriter, r *http.Request, params generated.ListProductsParams) {
	ctx := r.Context()

	var products []*ent.Product
	var err error

	if params.CategoryId != nil {
		products, err = h.products.GetByCategory(ctx, uuid.UUID(*params.CategoryId))
	} else {
		products, err = h.products.GetAll(ctx)
	}
	if err != nil {
		writeEntError(w, err)
		return
	}

	apiProducts := toAPIProducts(products)

	// Enrich with inventory stock levels.
	ids := make([]uuid.UUID, len(products))
	for i, p := range products {
		ids[i] = p.ID
	}
	if stocks, err := h.products.GetStockBatch(ctx, ids); err == nil {
		for i, p := range products {
			if s, ok := stocks[p.ID]; ok {
				apiProducts[i].Stock = &s
			}
		}
	}

	response.WriteJSON(w, http.StatusOK, generated.ProductList{
		Items: apiProducts,
	})
}

// CreateProduct creates a new product.
// (POST /products)
func (h *Handlers) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var body generated.ProductCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	pType := product.TypeSimple
	if body.Type != nil {
		pType = product.Type(*body.Type)
	}

	var jetonID *uuid.UUID
	if body.JetonId != nil {
		id := uuid.UUID(*body.JetonId)
		jetonID = &id
	}

	prod, err := h.products.Create(
		r.Context(),
		uuid.UUID(body.CategoryId),
		pType,
		body.Name,
		body.PriceCents,
		true, // isActive defaults to true
		body.Image,
		jetonID,
	)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, toAPIProduct(prod))
}

// GetProduct returns a single product by ID.
// (GET /products/{productId})
func (h *Handlers) GetProduct(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID) {
	ctx := r.Context()
	id := uuid.UUID(productId)
	prod, err := h.products.GetByID(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}
	ap := toAPIProduct(prod)
	if stock, err := h.products.GetStock(ctx, id); err == nil {
		s := int(stock)
		ap.Stock = &s
	}
	response.WriteJSON(w, http.StatusOK, ap)
}

// UpdateProduct partially updates a product.
// (PATCH /products/{productId})
func (h *Handlers) UpdateProduct(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID) {
	ctx := r.Context()
	id := uuid.UUID(productId)

	existing, err := h.products.GetByID(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}

	var body generated.ProductUpdate
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
	categoryID := existing.CategoryID
	if body.CategoryId != nil {
		categoryID = uuid.UUID(*body.CategoryId)
	}
	image := existing.Image
	if body.Image != nil {
		image = body.Image
	}
	jetonID := existing.JetonID
	if body.JetonId != nil {
		id := uuid.UUID(*body.JetonId)
		jetonID = &id
	}

	prod, err := h.products.Update(
		ctx,
		id,
		categoryID,
		existing.Type,
		name,
		priceCents,
		isActive,
		image,
		jetonID,
	)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIProduct(prod))
}

// DeleteProduct removes a product.
// (DELETE /products/{productId})
func (h *Handlers) DeleteProduct(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID) {
	if err := h.products.Delete(r.Context(), uuid.UUID(productId)); err != nil {
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetProductInventory returns the current inventory for a product.
// (GET /products/{productId}/inventory)
func (h *Handlers) GetProductInventory(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID) {
	ctx := r.Context()
	id := uuid.UUID(productId)

	// Verify product exists.
	_, err := h.products.GetByID(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}

	stock, err := h.products.GetStock(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, generated.Inventory{
		ProductId: openapi_types.UUID(id),
		Quantity:  int(stock),
	})
}

// GetProductInventoryHistory returns paginated inventory ledger entries for a product.
// (GET /products/{productId}/inventory/history)
func (h *Handlers) GetProductInventoryHistory(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID, params generated.GetProductInventoryHistoryParams) {
	ctx := r.Context()
	id := uuid.UUID(productId)

	// Verify product exists.
	_, err := h.products.GetByID(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}

	limit := 50
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	entries, err := h.products.ListInventoryHistory(ctx, id, limit, offset)
	if err != nil {
		writeEntError(w, err)
		return
	}

	items := make([]generated.InventoryLedgerEntry, 0, len(entries))
	for _, e := range entries {
		items = append(items, toAPIInventoryLedgerEntry(e))
	}
	response.WriteJSON(w, http.StatusOK, generated.InventoryLedgerList{Items: items})
}

// AdjustProductInventory adjusts the inventory for a product.
// (PATCH /products/{productId}/inventory)
func (h *Handlers) AdjustProductInventory(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID) {
	var body generated.InventoryAdjustment
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	ctx := r.Context()
	id := uuid.UUID(productId)

	// Verify product exists.
	_, err := h.products.GetByID(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}

	if err := h.products.AdjustStock(ctx, id, int64(body.Delta), string(body.Reason)); err != nil {
		writeEntError(w, err)
		return
	}

	stock, err := h.products.GetStock(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, generated.Inventory{
		ProductId: openapi_types.UUID(id),
		Quantity:  int(stock),
	})
}

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func (h *Handlers) UploadProductImage(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID) {
	if h.blobStore == nil {
		writeError(w, http.StatusNotImplemented, "not_configured", "Image uploads not configured")
		return
	}

	const maxSize = 5 << 20 // 5 MB
	if err := r.ParseMultipartForm(maxSize); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "File too large or invalid multipart form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Missing file field")
		return
	}
	defer func() { _ = file.Close() }()

	buf := make([]byte, 512)
	n, err := io.ReadAtLeast(file, buf, 1)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Could not read file")
		return
	}
	contentType := http.DetectContentType(buf[:n])
	if !allowedImageTypes[contentType] {
		writeError(w, http.StatusBadRequest, "invalid_file_type", "Only JPEG, PNG, and WebP images are allowed")
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.")
		return
	}

	ctx := r.Context()
	id := uuid.UUID(productId)
	existing, err := h.products.GetByID(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}

	if existing.Image != nil && *existing.Image != "" {
		_ = h.blobStore.Delete(ctx, *existing.Image)
	}

	var allData bytes.Buffer
	if _, err := io.Copy(&allData, file); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.")
		return
	}

	reader := bytes.NewReader(allData.Bytes())
	url, err := h.blobStore.Upload(ctx, reader, contentType)
	if err != nil {
		h.logger.Error("failed to upload image", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "upload_failed", "Failed to upload image")
		return
	}

	_, err = h.products.Update(ctx, id, existing.CategoryID, existing.Type, existing.Name, existing.PriceCents, existing.IsActive, &url, existing.JetonID)
	if err != nil {
		writeEntError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, generated.ProductImageResponse{ImageUrl: url})
}

func (h *Handlers) DeleteProductImage(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID) {
	ctx := r.Context()
	id := uuid.UUID(productId)

	existing, err := h.products.GetByID(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}

	if existing.Image != nil && *existing.Image != "" {
		if h.blobStore != nil {
			_ = h.blobStore.Delete(ctx, *existing.Image)
		}
	}

	_, err = h.products.Update(ctx, id, existing.CategoryID, existing.Type, existing.Name, existing.PriceCents, existing.IsActive, nil, existing.JetonID)
	if err != nil {
		writeEntError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
