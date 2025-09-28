package repository

import (
    "backend/internal/database"
    "backend/internal/domain"
    "context"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type OrderRepository interface {
    Create(ctx context.Context, o *domain.Order) (primitive.ObjectID, error)
    SetStripeSession(ctx context.Context, id primitive.ObjectID, sessionID string) error
    UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.OrderStatus) error
    FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error)
}

type orderRepository struct {
    collection *mongo.Collection
}

func NewOrderRepository(db *database.MongoDB) OrderRepository {
    return &orderRepository{
        collection: db.Database.Collection(database.OrdersCollection),
    }
}

func (r *orderRepository) Create(ctx context.Context, o *domain.Order) (primitive.ObjectID, error) {
    if o.ID.IsZero() {
        o.ID = primitive.NewObjectID()
    }
    now := time.Now().UTC()
    if o.CreatedAt.IsZero() { o.CreatedAt = now }
    o.UpdatedAt = now
    _, err := r.collection.InsertOne(ctx, o)
    if err != nil {
        return primitive.NilObjectID, err
    }
    return o.ID, nil
}

func (r *orderRepository) SetStripeSession(ctx context.Context, id primitive.ObjectID, sessionID string) error {
    _, err := r.collection.UpdateByID(ctx, id, bson.M{
        "$set": bson.M{
            "stripe_session_id": sessionID,
            "updated_at":       time.Now().UTC(),
        },
    })
    return err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.OrderStatus) error {
    _, err := r.collection.UpdateByID(ctx, id, bson.M{
        "$set": bson.M{
            "status":     status,
            "updated_at": time.Now().UTC(),
        },
    })
    return err
}

func (r *orderRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error) {
    var o domain.Order
    err := r.collection.FindOne(ctx, bson.M{"_id": id}, options.FindOne()).Decode(&o)
    if err != nil {
        return nil, err
    }
    return &o, nil
}
