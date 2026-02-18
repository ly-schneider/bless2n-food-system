package repository

import (
	"context"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/orderline"
	"backend/internal/generated/ent/orderlineredemption"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

type OrderLineRepository interface {
	Create(ctx context.Context, orderID uuid.UUID, lineType orderline.LineType, productID uuid.UUID, title string, quantity int, unitPriceCents int64, parentLineID, menuSlotID *uuid.UUID, menuSlotName *string) (*ent.OrderLine, error)
	CreateBatch(ctx context.Context, lines []OrderLineCreateParams) ([]*ent.OrderLine, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.OrderLine, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*ent.OrderLine, error)
	GetUnredeemed(ctx context.Context, orderID uuid.UUID) ([]*ent.OrderLine, error)
	Update(ctx context.Context, id, orderID uuid.UUID, lineType orderline.LineType, productID uuid.UUID, title string, quantity int, unitPriceCents int64, parentLineID, menuSlotID *uuid.UUID, menuSlotName *string) (*ent.OrderLine, error)
	GetByOrderAndProductIDs(ctx context.Context, orderID uuid.UUID, productIDs []uuid.UUID) ([]*ent.OrderLine, error)
}

// OrderLineCreateParams holds the parameters needed to create an order line in a batch.
type OrderLineCreateParams struct {
	ID             *uuid.UUID
	OrderID        uuid.UUID
	LineType       orderline.LineType
	ProductID      uuid.UUID
	Title          string
	Quantity       int
	UnitPriceCents int64
	ParentLineID   *uuid.UUID
	MenuSlotID     *uuid.UUID
	MenuSlotName   *string
}

type orderLineRepo struct {
	client *ent.Client
}

func NewOrderLineRepository(client *ent.Client) OrderLineRepository {
	return &orderLineRepo{client: client}
}

func (r *orderLineRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *orderLineRepo) Create(ctx context.Context, orderID uuid.UUID, lineType orderline.LineType, productID uuid.UUID, title string, quantity int, unitPriceCents int64, parentLineID, menuSlotID *uuid.UUID, menuSlotName *string) (*ent.OrderLine, error) {
	builder := r.ec(ctx).OrderLine.Create().
		SetOrderID(orderID).
		SetLineType(lineType).
		SetProductID(productID).
		SetTitle(title).
		SetQuantity(quantity).
		SetUnitPriceCents(unitPriceCents)
	if parentLineID != nil {
		builder.SetParentLineID(*parentLineID)
	}
	if menuSlotID != nil {
		builder.SetMenuSlotID(*menuSlotID)
	}
	if menuSlotName != nil {
		builder.SetMenuSlotName(*menuSlotName)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *orderLineRepo) CreateBatch(ctx context.Context, lines []OrderLineCreateParams) ([]*ent.OrderLine, error) {
	if len(lines) == 0 {
		return nil, nil
	}
	builders := make([]*ent.OrderLineCreate, len(lines))
	for i, line := range lines {
		b := r.ec(ctx).OrderLine.Create().
			SetOrderID(line.OrderID).
			SetLineType(line.LineType).
			SetProductID(line.ProductID).
			SetTitle(line.Title).
			SetQuantity(line.Quantity).
			SetUnitPriceCents(line.UnitPriceCents)
		if line.ID != nil {
			b.SetID(*line.ID)
		}
		if line.ParentLineID != nil {
			b.SetParentLineID(*line.ParentLineID)
		}
		if line.MenuSlotID != nil {
			b.SetMenuSlotID(*line.MenuSlotID)
		}
		if line.MenuSlotName != nil {
			b.SetMenuSlotName(*line.MenuSlotName)
		}
		builders[i] = b
	}
	created, err := r.ec(ctx).OrderLine.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *orderLineRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.OrderLine, error) {
	e, err := r.ec(ctx).OrderLine.Query().
		Where(orderline.ID(id)).
		WithProduct().
		WithRedemption().
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *orderLineRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*ent.OrderLine, error) {
	rows, err := r.ec(ctx).OrderLine.Query().
		Where(orderline.OrderIDEQ(orderID)).
		WithProduct().
		WithRedemption().
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *orderLineRepo) GetUnredeemed(ctx context.Context, orderID uuid.UUID) ([]*ent.OrderLine, error) {
	// Find order lines that have no redemption record
	rows, err := r.ec(ctx).OrderLine.Query().
		Where(
			orderline.OrderIDEQ(orderID),
			orderline.Not(orderline.HasRedemption()),
		).
		WithProduct().
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *orderLineRepo) Update(ctx context.Context, id, orderID uuid.UUID, lineType orderline.LineType, productID uuid.UUID, title string, quantity int, unitPriceCents int64, parentLineID, menuSlotID *uuid.UUID, menuSlotName *string) (*ent.OrderLine, error) {
	builder := r.ec(ctx).OrderLine.UpdateOneID(id).
		SetOrderID(orderID).
		SetLineType(lineType).
		SetProductID(productID).
		SetTitle(title).
		SetQuantity(quantity).
		SetUnitPriceCents(unitPriceCents)
	if parentLineID != nil {
		builder.SetParentLineID(*parentLineID)
	} else {
		builder.ClearParentLineID()
	}
	if menuSlotID != nil {
		builder.SetMenuSlotID(*menuSlotID)
	} else {
		builder.ClearMenuSlotID()
	}
	if menuSlotName != nil {
		builder.SetMenuSlotName(*menuSlotName)
	} else {
		builder.ClearMenuSlotName()
	}
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return updated, nil
}

func (r *orderLineRepo) GetByOrderAndProductIDs(ctx context.Context, orderID uuid.UUID, productIDs []uuid.UUID) ([]*ent.OrderLine, error) {
	if len(productIDs) == 0 {
		return []*ent.OrderLine{}, nil
	}
	rows, err := r.ec(ctx).OrderLine.Query().
		Where(
			orderline.OrderIDEQ(orderID),
			orderline.ProductIDIn(productIDs...),
		).
		WithProduct().
		WithRedemption().
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

// Import anchor for packages used in predicates
var _ = orderlineredemption.Table
var _ = sql.EQ
