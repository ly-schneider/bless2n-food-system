package repository

import (
	"context"
	"time"

	"backend/internal/database"
	"backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error)
	GetByCustomerID(ctx context.Context, customerID primitive.ObjectID, limit, offset int) ([]*domain.Order, error)
	GetByContactEmail(ctx context.Context, email string, limit, offset int) ([]*domain.Order, error)
	GetByStatus(ctx context.Context, status domain.OrderStatus, limit, offset int) ([]*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.OrderStatus) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, limit, offset int) ([]*domain.Order, error)
	GetRecent(ctx context.Context, limit int) ([]*domain.Order, error)
}

type orderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *database.MongoDB) OrderRepository {
	return &orderRepository{
		collection: db.Database.Collection(database.OrdersCollection),
	}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = domain.OrderStatusPending

	_, err := r.collection.InsertOne(ctx, order)
	return err
}

func (r *orderRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error) {
	var order domain.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByCustomerID(ctx context.Context, customerID primitive.ObjectID, limit, offset int) ([]*domain.Order, error) {
	filter := bson.M{"customer_id": customerID}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var order domain.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, cursor.Err()
}

func (r *orderRepository) GetByContactEmail(ctx context.Context, email string, limit, offset int) ([]*domain.Order, error) {
	filter := bson.M{"contact_email": email}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var order domain.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, cursor.Err()
}

func (r *orderRepository) GetByStatus(ctx context.Context, status domain.OrderStatus, limit, offset int) ([]*domain.Order, error) {
	filter := bson.M{"status": status}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var order domain.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, cursor.Err()
}

func (r *orderRepository) Update(ctx context.Context, order *domain.Order) error {
	order.UpdatedAt = time.Now()
	filter := bson.M{"_id": order.ID}
	update := bson.M{"$set": order}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.OrderStatus) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *orderRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *orderRepository) List(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var order domain.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, cursor.Err()
}

func (r *orderRepository) GetRecent(ctx context.Context, limit int) ([]*domain.Order, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var order domain.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, cursor.Err()
}
