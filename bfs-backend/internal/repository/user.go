package repository

import (
    "backend/internal/database"
    "backend/internal/domain"
    "context"
    "strings"
    "time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	UpsertCustomerByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateStripeCustomerID(ctx context.Context, id primitive.ObjectID, stripeID string) error
	UpdateEmail(ctx context.Context, id primitive.ObjectID, newEmail string, isVerified bool) error
	UpdateNames(ctx context.Context, id primitive.ObjectID, firstName, lastName *string) error
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, limit, offset int) ([]*domain.User, int64, error)
	UpsertByEmailWithRole(ctx context.Context, email string, role domain.UserRole, isVerified bool, firstName, lastName *string) (*domain.User, error)
	UpdateRole(ctx context.Context, id primitive.ObjectID, role domain.UserRole) error
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *database.MongoDB) UserRepository {
	return &userRepository{
		collection: db.Database.Collection(database.UsersCollection),
	}
}

// normalizeEmail trims spaces and lowercases the email for consistent storage and lookup.
func normalizeEmail(email string) string {
    return strings.ToLower(strings.TrimSpace(email))
}

func (r *userRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var u domain.User
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    email = normalizeEmail(email)
    var u domain.User
    if err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&u); err != nil {
        return nil, err
    }
    return &u, nil
}

// UpsertCustomerByEmail finds or creates a customer user by email.
func (r *userRepository) UpsertCustomerByEmail(ctx context.Context, email string) (*domain.User, error) {
    now := time.Now().UTC()
    email = normalizeEmail(email)
    // Try find first
    if u, err := r.FindByEmail(ctx, email); err == nil {
        return u, nil
    }
    u := &domain.User{
        ID:         primitive.NewObjectID(),
        Email:      email,
        Role:       domain.UserRoleCustomer,
        IsVerified: true,
        CreatedAt:  now,
        UpdatedAt:  now,
    }
	if _, err := r.collection.InsertOne(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) UpdateStripeCustomerID(ctx context.Context, id primitive.ObjectID, stripeID string) error {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"stripe_customer_id": stripeID,
			"updated_at":         now,
		},
	}
	_, err := r.collection.UpdateByID(ctx, id, update)
	return err
}

func (r *userRepository) UpdateEmail(ctx context.Context, id primitive.ObjectID, newEmail string, isVerified bool) error {
    now := time.Now().UTC()
    newEmail = normalizeEmail(newEmail)
    update := bson.M{
        "$set": bson.M{
            "email":       newEmail,
            "is_verified": isVerified,
            "updated_at":  now,
        },
    }
    _, err := r.collection.UpdateByID(ctx, id, update)
    return err
}

func (r *userRepository) UpdateNames(ctx context.Context, id primitive.ObjectID, firstName, lastName *string) error {
	now := time.Now().UTC()
	set := bson.M{
		"updated_at": now,
	}
	if firstName != nil {
		set["first_name"] = *firstName
	}
	if lastName != nil {
		set["last_name"] = *lastName
	}
	if len(set) == 1 { // only updated_at
		return nil
	}
	update := bson.M{"$set": set}
	_, err := r.collection.UpdateByID(ctx, id, update)
	return err
}

func (r *userRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, int64, error) {
	total, err := r.collection.CountDocuments(ctx, bson.M{})
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
	cur, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var users []*domain.User
	for cur.Next(ctx) {
		var u domain.User
		if err := cur.Decode(&u); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}
	if err := cur.Err(); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) UpsertByEmailWithRole(ctx context.Context, email string, role domain.UserRole, isVerified bool, firstName, lastName *string) (*domain.User, error) {
    now := time.Now().UTC()
    email = normalizeEmail(email)
    // Try find first
    if u, err := r.FindByEmail(ctx, email); err == nil && u != nil {
        // Upgrade role if needed, update verified and optional names
        set := bson.M{"updated_at": now}
		if u.Role != role {
			set["role"] = role
		}
		if isVerified && !u.IsVerified {
			set["is_verified"] = true
		}
		if firstName != nil {
			set["first_name"] = *firstName
		}
		if lastName != nil {
			set["last_name"] = *lastName
		}
		if len(set) > 1 {
			_, _ = r.collection.UpdateByID(ctx, u.ID, bson.M{"$set": set})
		}
		// Refresh
		return r.FindByEmail(ctx, email)
	}
	// Insert new
	u := &domain.User{
		ID:    primitive.NewObjectID(),
		Email: email,
		FirstName: func() string {
			if firstName != nil {
				return *firstName
			}
			return ""
		}(),
		LastName: func() string {
			if lastName != nil {
				return *lastName
			}
			return ""
		}(),
		Role:       role,
		IsVerified: isVerified,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if _, err := r.collection.InsertOne(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) UpdateRole(ctx context.Context, id primitive.ObjectID, role domain.UserRole) error {
	now := time.Now().UTC()
	_, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"role": role, "updated_at": now}})
	return err
}
