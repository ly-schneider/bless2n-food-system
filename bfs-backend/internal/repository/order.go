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
    SetStripePaymentIntent(ctx context.Context, id primitive.ObjectID, paymentIntentID string, customerID *string, receiptEmail *string) error
    SetStripePaymentSuccess(ctx context.Context, id primitive.ObjectID, paymentIntentID string, chargeID *string, customerID *string, receiptEmail *string) error
    SetPaymentAttemptID(ctx context.Context, id primitive.ObjectID, attemptID string) error
    UpdateStatusAndContact(ctx context.Context, id primitive.ObjectID, status domain.OrderStatus, contactEmail *string) error
    FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error)
    DeleteIfPending(ctx context.Context, id primitive.ObjectID) (bool, error)
    FindPendingByStripeSessionID(ctx context.Context, sessionID string) (*domain.Order, error)
    FindPendingByAttemptID(ctx context.Context, attemptID string) (*domain.Order, error)
    DeletePendingByAttemptIDExcept(ctx context.Context, attemptID string, except primitive.ObjectID) (int64, error)
    // ListByCustomerID returns orders for a given customer with pagination
    ListByCustomerID(ctx context.Context, customerID primitive.ObjectID, limit, offset int) ([]*domain.Order, int64, error)
    // ListAdmin returns orders with admin filters
    ListAdmin(ctx context.Context, status *domain.OrderStatus, from, to *time.Time, q *string, limit, offset int) ([]*domain.Order, int64, error)
}

type orderRepository struct {
    collection *mongo.Collection
}

func NewOrderRepository(db *database.MongoDB) OrderRepository {
    coll := db.Database.Collection(database.OrdersCollection)
    // Ensure partial unique index on payment_attempt_id for pending orders to avoid duplicates
    // db.orders.createIndex({ payment_attempt_id: 1 }, { unique: true, partialFilterExpression: { status: 'pending', payment_attempt_id: { $exists: true } } })
    _, _ = coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
        Keys: bson.D{{Key: "payment_attempt_id", Value: 1}},
        Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{
            "status":              domain.OrderStatusPending,
            "payment_attempt_id": bson.M{"$exists": true},
        }),
    })
    return &orderRepository{ collection: coll }
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

func (r *orderRepository) SetStripePaymentIntent(ctx context.Context, id primitive.ObjectID, paymentIntentID string, customerID *string, receiptEmail *string) error {
    set := bson.M{
        "stripe_payment_intent_id": paymentIntentID,
        "updated_at":               time.Now().UTC(),
    }
    if customerID != nil && *customerID != "" {
        set["stripe_customer_id"] = *customerID
    }
    if receiptEmail != nil && *receiptEmail != "" {
        set["contact_email"] = *receiptEmail
    }
    _, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": set})
    return err
}

func (r *orderRepository) SetStripePaymentSuccess(ctx context.Context, id primitive.ObjectID, paymentIntentID string, chargeID *string, customerID *string, receiptEmail *string) error {
    set := bson.M{
        "status":                   domain.OrderStatusPaid,
        "stripe_payment_intent_id": paymentIntentID,
        "updated_at":               time.Now().UTC(),
    }
    if chargeID != nil && *chargeID != "" {
        set["stripe_charge_id"] = *chargeID
    }
    if customerID != nil && *customerID != "" {
        set["stripe_customer_id"] = *customerID
    }
    if receiptEmail != nil && *receiptEmail != "" {
        set["contact_email"] = *receiptEmail
    }
    _, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": set})
    return err
}

func (r *orderRepository) SetPaymentAttemptID(ctx context.Context, id primitive.ObjectID, attemptID string) error {
    if attemptID == "" { return nil }
    set := bson.M{
        "payment_attempt_id": attemptID,
        "updated_at":         time.Now().UTC(),
    }
    _, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": set})
    return err
}

func (r *orderRepository) UpdateStatusAndContact(ctx context.Context, id primitive.ObjectID, status domain.OrderStatus, contactEmail *string) error {
    set := bson.M{
        "status":     status,
        "updated_at": time.Now().UTC(),
    }
    if contactEmail != nil && *contactEmail != "" {
        set["contact_email"] = *contactEmail
    }
    _, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": set})
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

func (r *orderRepository) DeleteIfPending(ctx context.Context, id primitive.ObjectID) (bool, error) {
    res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id, "status": domain.OrderStatusPending})
    if err != nil {
        return false, err
    }
    return res.DeletedCount > 0, nil
}

func (r *orderRepository) FindPendingByStripeSessionID(ctx context.Context, sessionID string) (*domain.Order, error) {
    var o domain.Order
    err := r.collection.FindOne(ctx, bson.M{"stripe_session_id": sessionID, "status": domain.OrderStatusPending}).Decode(&o)
    if err != nil {
        return nil, err
    }
    return &o, nil
}

func (r *orderRepository) FindPendingByAttemptID(ctx context.Context, attemptID string) (*domain.Order, error) {
    var o domain.Order
    err := r.collection.FindOne(ctx, bson.M{"payment_attempt_id": attemptID, "status": domain.OrderStatusPending}).Decode(&o)
    if err != nil {
        return nil, err
    }
    return &o, nil
}

func (r *orderRepository) DeletePendingByAttemptIDExcept(ctx context.Context, attemptID string, except primitive.ObjectID) (int64, error) {
    if attemptID == "" { return 0, nil }
    res, err := r.collection.DeleteMany(ctx, bson.M{
        "payment_attempt_id": attemptID,
        "status":             domain.OrderStatusPending,
        "_id":                bson.M{"$ne": except},
    })
    if err != nil { return 0, err }
    return res.DeletedCount, nil
}

func (r *orderRepository) ListByCustomerID(ctx context.Context, customerID primitive.ObjectID, limit, offset int) ([]*domain.Order, int64, error) {
    filter := bson.M{"customer_id": customerID}
    // Count total matching for pagination
    total, err := r.collection.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }

    opts := options.Find().SetSort(bson.M{"created_at": -1})
    if limit > 0 {
        opts.SetLimit(int64(limit))
    }
    if offset > 0 {
        opts.SetSkip(int64(offset))
    }

    cur, err := r.collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, 0, err
    }
    defer func() { _ = cur.Close(ctx) }()

    var orders []*domain.Order
    for cur.Next(ctx) {
        var o domain.Order
        if err := cur.Decode(&o); err != nil {
            return nil, 0, err
        }
        orders = append(orders, &o)
    }
    if err := cur.Err(); err != nil {
        return nil, 0, err
    }
    return orders, total, nil
}

func (r *orderRepository) ListAdmin(ctx context.Context, status *domain.OrderStatus, from, to *time.Time, q *string, limit, offset int) ([]*domain.Order, int64, error) {
    filter := bson.M{}
    if status != nil { filter["status"] = *status }
    if from != nil || to != nil {
        created := bson.M{}
        if from != nil { created["$gte"] = *from }
        if to != nil { created["$lt"] = *to }
        filter["created_at"] = created
    }
    if q != nil && *q != "" {
        // search by id prefix or contact email
        or := []bson.M{}
        // try ObjectID
        if oid, err := primitive.ObjectIDFromHex(*q); err == nil {
            or = append(or, bson.M{"_id": oid})
        }
        or = append(or, bson.M{"contact_email": bson.M{"$regex": *q, "$options": "i"}})
        filter["$or"] = or
    }
    total, err := r.collection.CountDocuments(ctx, filter)
    if err != nil { return nil, 0, err }
    opts := options.Find().SetSort(bson.M{"created_at": -1})
    if limit > 0 { opts.SetLimit(int64(limit)) }
    if offset > 0 { opts.SetSkip(int64(offset)) }
    cur, err := r.collection.Find(ctx, filter, opts)
    if err != nil { return nil, 0, err }
    defer func() { _ = cur.Close(ctx) }()
    var orders []*domain.Order
    for cur.Next(ctx) {
        var o domain.Order
        if err := cur.Decode(&o); err != nil { return nil, 0, err }
        orders = append(orders, &o)
    }
    if err := cur.Err(); err != nil { return nil, 0, err }
    return orders, total, nil
}
