package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PosDeviceRepository interface {
	UpsertPendingByToken(ctx context.Context, name, token string) (*domain.PosDevice, error)
	FindByToken(ctx context.Context, token string) (*domain.PosDevice, error)
	ApproveByToken(ctx context.Context, token string) error
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.PosDevice, error)
	List(ctx context.Context) ([]*domain.PosDevice, error)
	UpdateConfig(ctx context.Context, id bson.ObjectID, cardCapable *bool, printerMAC *string, printerUUID *string) error
}

type posDeviceRepository struct {
	collection *mongo.Collection
}

func NewPosDeviceRepository(db *database.MongoDB) PosDeviceRepository {
	return &posDeviceRepository{collection: db.Database.Collection(database.PosDevicesCollection)}
}

func (r *posDeviceRepository) UpsertPendingByToken(ctx context.Context, name, token string) (*domain.PosDevice, error) {
	now := time.Now().UTC()
	filter := bson.M{"device_token": token}
	update := bson.M{"$setOnInsert": bson.M{
		"_id":          bson.NewObjectID(),
		"name":         name,
		"device_token": token,
		"approved":     false,
		"created_at":   now,
	}, "$set": bson.M{"updated_at": now}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var d domain.PosDevice
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&d); err != nil {
		if err == mongo.ErrNoDocuments {
			return r.FindByToken(ctx, token)
		}
		return nil, err
	}
	return &d, nil
}

func (r *posDeviceRepository) FindByToken(ctx context.Context, token string) (*domain.PosDevice, error) {
	var d domain.PosDevice
	if err := r.collection.FindOne(ctx, bson.M{"device_token": token}).Decode(&d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *posDeviceRepository) ApproveByToken(ctx context.Context, token string) error {
	now := time.Now().UTC()
	_, err := r.collection.UpdateOne(ctx, bson.M{"device_token": token}, bson.M{"$set": bson.M{"approved": true, "approved_at": now, "updated_at": now}})
	return err
}

func (r *posDeviceRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.PosDevice, error) {
	var d domain.PosDevice
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *posDeviceRepository) List(ctx context.Context) ([]*domain.PosDevice, error) {
	cur, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	out := make([]*domain.PosDevice, 0)
	for cur.Next(ctx) {
		var it domain.PosDevice
		if err := cur.Decode(&it); err != nil {
			return nil, err
		}
		out = append(out, &it)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *posDeviceRepository) UpdateConfig(ctx context.Context, id bson.ObjectID, cardCapable *bool, printerMAC *string, printerUUID *string) error {
	set := bson.M{"updated_at": time.Now().UTC()}
	if cardCapable != nil {
		set["card_capable"] = *cardCapable
	}
	if printerMAC != nil {
		set["printer_mac"] = *printerMAC
	}
	if printerUUID != nil {
		set["printer_uuid"] = *printerUUID
	}
	_, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": set})
	return err
}
