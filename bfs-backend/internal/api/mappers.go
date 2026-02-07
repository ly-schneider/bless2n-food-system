package api

import (
	"time"

	"backend/internal/generated/api/generated"
	"backend/internal/generated/ent"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ---------------------------------------------------------------------------
// Generic helpers
// ---------------------------------------------------------------------------

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T { return &v }

// derefStr safely dereferences a *string, returning "" if nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ---------------------------------------------------------------------------
// Category
// ---------------------------------------------------------------------------

func toAPICategory(e *ent.Category) generated.Category {
	return generated.Category{
		Id:        e.ID,
		Name:      e.Name,
		IsActive:  e.IsActive,
		Position:  e.Position,
		CreatedAt: ptr(e.CreatedAt),
		UpdatedAt: ptr(e.UpdatedAt),
	}
}

func toAPICategories(rows []*ent.Category) []generated.Category {
	out := make([]generated.Category, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAPICategory(r))
	}
	return out
}

func toAPICategorySummary(e *ent.Category) generated.CategorySummary {
	return generated.CategorySummary{
		Id:       ptr(openapi_types.UUID(e.ID)),
		Name:     ptr(e.Name),
		Position: ptr(e.Position),
	}
}

// ---------------------------------------------------------------------------
// Product
// ---------------------------------------------------------------------------

func toAPIProduct(e *ent.Product) generated.Product {
	p := generated.Product{
		Id:         e.ID,
		CategoryId: e.CategoryID,
		Type:       generated.ProductType(e.Type),
		Name:       e.Name,
		Image:      e.Image,
		PriceCents: e.PriceCents,
		JetonId:    (*openapi_types.UUID)(e.JetonID),
		IsActive:   e.IsActive,
		CreatedAt:  ptr(e.CreatedAt),
		UpdatedAt:  ptr(e.UpdatedAt),
	}

	// Map Category edge if loaded.
	if e.Edges.Category != nil {
		cs := toAPICategorySummary(e.Edges.Category)
		p.Category = &cs
	}

	// Map Jeton edge if loaded.
	if e.Edges.Jeton != nil {
		js := toAPIJetonSummary(e.Edges.Jeton)
		p.Jeton = &js
	}

	// Map MenuSlots edge if loaded.
	if slots, err := e.Edges.MenuSlotsOrErr(); err == nil && len(slots) > 0 {
		summaries := make([]generated.MenuSlotSummary, 0, len(slots))
		for _, s := range slots {
			summaries = append(summaries, toAPIMenuSlotSummary(s))
		}
		p.MenuSlots = &summaries
	}

	return p
}

func toAPIProducts(rows []*ent.Product) []generated.Product {
	out := make([]generated.Product, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAPIProduct(r))
	}
	return out
}

// ---------------------------------------------------------------------------
// Jeton
// ---------------------------------------------------------------------------

func toAPIJeton(e *ent.Jeton) generated.Jeton {
	return generated.Jeton{
		Id:        e.ID,
		Name:      e.Name,
		Color:     e.Color,
		CreatedAt: ptr(e.CreatedAt),
		UpdatedAt: ptr(e.UpdatedAt),
	}
}

func toAPIJetons(rows []*ent.Jeton) []generated.Jeton {
	out := make([]generated.Jeton, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAPIJeton(r))
	}
	return out
}

func toAPIJetonSummary(e *ent.Jeton) generated.JetonSummary {
	return generated.JetonSummary{
		Id:       ptr(openapi_types.UUID(e.ID)),
		Name:     ptr(e.Name),
		Color: ptr(e.Color),
	}
}

// ---------------------------------------------------------------------------
// Order
// ---------------------------------------------------------------------------

func toAPIOrder(e *ent.Order) generated.Order {
	o := generated.Order{
		Id:                   e.ID,
		CustomerId:           e.CustomerID,
		TotalCents:           e.TotalCents,
		Status:               generated.OrderStatus(e.Status),
		Origin:               generated.OrderOrigin(e.Origin),
		CreatedAt:            e.CreatedAt,
		UpdatedAt:            ptr(e.UpdatedAt),
		PaymentAttemptId:     e.PaymentAttemptID,
		PayrexxGatewayId:     e.PayrexxGatewayID,
		PayrexxTransactionId: e.PayrexxTransactionID,
	}

	// ContactEmail: *string in ent -> *openapi_types.Email in API
	if e.ContactEmail != nil {
		o.ContactEmail = (*openapi_types.Email)(e.ContactEmail)
	}

	// Map Lines edge if loaded.
	if lines, err := e.Edges.LinesOrErr(); err == nil {
		apiLines := make([]generated.OrderLine, 0, len(lines))
		for _, l := range lines {
			apiLines = append(apiLines, toAPIOrderLine(l))
		}
		o.Lines = &apiLines
	}

	// Map Payments edge if loaded.
	if payments, err := e.Edges.PaymentsOrErr(); err == nil {
		apiPayments := make([]generated.OrderPaymentSummary, 0, len(payments))
		for _, p := range payments {
			apiPayments = append(apiPayments, toAPIOrderPaymentSummary(p))
		}
		o.Payments = &apiPayments
	}

	return o
}

// ---------------------------------------------------------------------------
// OrderLine
// ---------------------------------------------------------------------------

func toAPIOrderLine(e *ent.OrderLine) generated.OrderLine {
	ol := generated.OrderLine{
		Id:             e.ID,
		OrderId:        e.OrderID,
		LineType:       generated.OrderItemType(e.LineType),
		ProductId:      e.ProductID,
		Title:          e.Title,
		Quantity:       e.Quantity,
		UnitPriceCents: e.UnitPriceCents,
		ParentLineId:   (*openapi_types.UUID)(e.ParentLineID),
		MenuSlotId:     (*openapi_types.UUID)(e.MenuSlotID),
		MenuSlotName:   e.MenuSlotName,
	}

	if e.Edges.Product != nil {
		ol.ProductImage = e.Edges.Product.Image
	}

	// Map Redemption edge if loaded.
	if e.Edges.Redemption != nil {
		r := toAPIOrderLineRedemption(e.Edges.Redemption)
		ol.Redemption = &r
	}

	// Map ChildLines edge if loaded.
	if children, err := e.Edges.ChildLinesOrErr(); err == nil && len(children) > 0 {
		childLines := make([]generated.OrderLine, 0, len(children))
		for _, c := range children {
			childLines = append(childLines, toAPIOrderLine(c))
		}
		ol.ChildLines = &childLines
	}

	return ol
}

// ---------------------------------------------------------------------------
// OrderLineRedemption
// ---------------------------------------------------------------------------

func toAPIOrderLineRedemption(e *ent.OrderLineRedemption) generated.OrderLineRedemption {
	return generated.OrderLineRedemption{
		Id:          e.ID,
		OrderLineId: e.OrderLineID,
		RedeemedAt:  e.RedeemedAt,
	}
}

// ---------------------------------------------------------------------------
// OrderPaymentSummary
// ---------------------------------------------------------------------------

func toAPIOrderPaymentSummary(e *ent.OrderPayment) generated.OrderPaymentSummary {
	method := generated.OrderPaymentSummaryMethod(e.Method)
	return generated.OrderPaymentSummary{
		Id:          ptr(openapi_types.UUID(e.ID)),
		Method:      &method,
		AmountCents: ptr(e.AmountCents),
		PaidAt:      ptr(e.PaidAt),
	}
}

// ---------------------------------------------------------------------------
// Device (from ent.Device entity)
// ---------------------------------------------------------------------------

func toAPIDevice(e *ent.Device) generated.Device {
	return generated.Device{
		Id:        e.ID,
		Name:      e.Name,
		Model:     e.Model,
		Os:        e.Os,
		DeviceKey: ptr(e.DeviceKey),
		Type:      generated.DeviceType(e.Type),
		Status:    generated.DeviceStatus(e.Status),
		DecidedBy: e.DecidedBy,
		DecidedAt: e.DecidedAt,
		ExpiresAt: e.ExpiresAt,
		CreatedAt: ptr(e.CreatedAt),
		UpdatedAt: ptr(e.UpdatedAt),
	}
}

func toAPIDevices(rows []*ent.Device) []generated.Device {
	out := make([]generated.Device, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAPIDevice(r))
	}
	return out
}

// ---------------------------------------------------------------------------
// Invite (from ent.AdminInvite)
// ---------------------------------------------------------------------------

func toAPIInvite(e *ent.AdminInvite) generated.Invite {
	inv := generated.Invite{
		Id:              e.ID,
		InvitedByUserId: e.InvitedByUserID,
		InviteeEmail:    openapi_types.Email(e.InviteeEmail),
		Status:          generated.InviteStatus(e.Status),
		ExpiresAt:       e.ExpiresAt,
		UsedAt:          e.UsedAt,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       ptr(e.UpdatedAt),
	}

	if e.Edges.Inviter != nil {
		us := toAPIUserSummary(e.Edges.Inviter)
		inv.Inviter = &us
	}

	return inv
}

func toAPIInvites(rows []*ent.AdminInvite) []generated.Invite {
	out := make([]generated.Invite, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAPIInvite(r))
	}
	return out
}

// ---------------------------------------------------------------------------
// User
// ---------------------------------------------------------------------------

func toAPIUser(e *ent.User) generated.User {
	u := generated.User{
		Id:            e.ID,
		Name:          e.Name,
		Image:         e.Image,
		EmailVerified: ptr(e.EmailVerified),
		IsAnonymous:   ptr(e.IsAnonymous),
		Role:          generated.UserRole(e.Role),
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     ptr(e.UpdatedAt),
	}

	// Email: *string in ent -> openapi_types.Email in API
	if e.Email != nil {
		u.Email = openapi_types.Email(*e.Email)
	}

	return u
}

func toAPIUserSummary(e *ent.User) generated.UserSummary {
	us := generated.UserSummary{
		Id:   ptr(e.ID),
		Name: e.Name,
	}
	if e.Email != nil {
		us.Email = (*openapi_types.Email)(e.Email)
	}
	return us
}

func toAPIUsers(rows []*ent.User) []generated.User {
	out := make([]generated.User, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAPIUser(r))
	}
	return out
}

// ---------------------------------------------------------------------------
// MenuSlot
// ---------------------------------------------------------------------------

func toAPIMenuSlot(e *ent.MenuSlot) generated.MenuSlot {
	ms := generated.MenuSlot{
		Id:            e.ID,
		MenuProductId: e.MenuProductID,
		Name:          e.Name,
		Sequence:      e.Sequence,
	}

	// Map Options edge if loaded.
	if opts, err := e.Edges.OptionsOrErr(); err == nil {
		apiOpts := make([]generated.MenuSlotOption, 0, len(opts))
		for _, o := range opts {
			apiOpts = append(apiOpts, toAPIMenuSlotOption(o))
		}
		ms.Options = &apiOpts
	}

	return ms
}

func toAPIMenuSlotSummary(e *ent.MenuSlot) generated.MenuSlotSummary {
	ms := generated.MenuSlotSummary{
		Id:       ptr(openapi_types.UUID(e.ID)),
		Name:     ptr(e.Name),
		Sequence: ptr(e.Sequence),
	}

	// Map Options edge if loaded, producing inline option summaries.
	if opts, err := e.Edges.OptionsOrErr(); err == nil {
		optSummaries := make([]struct {
			Image      *string              `json:"image"`
			Jeton      *generated.JetonSummary `json:"jeton,omitempty"`
			Name       *string              `json:"name,omitempty"`
			PriceCents *int64               `json:"priceCents,omitempty"`
			ProductId  *openapi_types.UUID  `json:"productId,omitempty"`
		}, 0, len(opts))
		for _, o := range opts {
			entry := struct {
				Image      *string              `json:"image"`
				Jeton      *generated.JetonSummary `json:"jeton,omitempty"`
				Name       *string              `json:"name,omitempty"`
				PriceCents *int64               `json:"priceCents,omitempty"`
				ProductId  *openapi_types.UUID  `json:"productId,omitempty"`
			}{
				ProductId: ptr(openapi_types.UUID(o.OptionProductID)),
			}
			if o.Edges.OptionProduct != nil {
				entry.Name = ptr(o.Edges.OptionProduct.Name)
				entry.PriceCents = ptr(o.Edges.OptionProduct.PriceCents)
				entry.Image = o.Edges.OptionProduct.Image
				if o.Edges.OptionProduct.Edges.Jeton != nil {
					js := toAPIJetonSummary(o.Edges.OptionProduct.Edges.Jeton)
					entry.Jeton = &js
				}
			}
			optSummaries = append(optSummaries, entry)
		}
		ms.Options = &optSummaries
	}

	return ms
}

// ---------------------------------------------------------------------------
// MenuSlotOption
// ---------------------------------------------------------------------------

func toAPIMenuSlotOption(e *ent.MenuSlotOption) generated.MenuSlotOption {
	mso := generated.MenuSlotOption{
		MenuSlotId:      e.MenuSlotID,
		OptionProductId: e.OptionProductID,
	}

	// Map OptionProduct edge if loaded.
	if e.Edges.OptionProduct != nil {
		p := e.Edges.OptionProduct
		mso.OptionProduct = &struct {
			Id         *openapi_types.UUID `json:"id,omitempty"`
			Image      *string             `json:"image"`
			Name       *string             `json:"name,omitempty"`
			PriceCents *int64              `json:"priceCents,omitempty"`
		}{
			Id:         ptr(openapi_types.UUID(p.ID)),
			Image:      p.Image,
			Name:       ptr(p.Name),
			PriceCents: ptr(p.PriceCents),
		}
	}

	return mso
}

// ---------------------------------------------------------------------------
// Station (from ent.Device with type=STATION)
// ---------------------------------------------------------------------------

func toAPIStation(e *ent.Device) generated.Station {
	s := generated.Station{
		Id:        e.ID,
		Name:      e.Name,
		Model:     e.Model,
		Os:        e.Os,
		Status:    generated.DeviceStatus(e.Status),
		DecidedAt: e.DecidedAt,
		DecidedBy: e.DecidedBy,
		CreatedAt: ptr(e.CreatedAt),
		UpdatedAt: ptr(e.UpdatedAt),
	}

	// Map DeviceProducts edge -> StationProduct list if loaded.
	if dps, err := e.Edges.DeviceProductsOrErr(); err == nil {
		products := make([]generated.StationProduct, 0, len(dps))
		for _, dp := range dps {
			sp := generated.StationProduct{
				DeviceId:  ptr(openapi_types.UUID(dp.DeviceID)),
				ProductId: ptr(openapi_types.UUID(dp.ProductID)),
			}
			// If the Product edge is loaded on the DeviceProduct, include details.
			if dp.Edges.Product != nil {
				sp.Name = ptr(dp.Edges.Product.Name)
				sp.PriceCents = ptr(dp.Edges.Product.PriceCents)
			}
			products = append(products, sp)
		}
		s.Products = &products
	}

	return s
}

func toAPIStations(rows []*ent.Device) []generated.Station {
	out := make([]generated.Station, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAPIStation(r))
	}
	return out
}

// ---------------------------------------------------------------------------
// InventoryLedger
// ---------------------------------------------------------------------------

func toAPIInventoryLedgerEntry(e *ent.InventoryLedger) generated.InventoryLedgerEntry {
	entry := generated.InventoryLedgerEntry{
		Id:        e.ID,
		ProductId: openapi_types.UUID(e.ProductID),
		Delta:     e.Delta,
		Reason:    generated.InventoryReason(e.Reason),
		CreatedAt: e.CreatedAt,
		CreatedBy: e.CreatedBy,
	}
	if e.OrderID != nil {
		entry.OrderId = (*openapi_types.UUID)(e.OrderID)
	}
	if e.OrderLineID != nil {
		entry.OrderLineId = (*openapi_types.UUID)(e.OrderLineID)
	}
	if e.DeviceID != nil {
		entry.DeviceId = (*openapi_types.UUID)(e.DeviceID)
	}
	return entry
}

// ---------------------------------------------------------------------------
// Suppress unused import warnings — these are used by mapper functions.
// ---------------------------------------------------------------------------

var _ time.Time
