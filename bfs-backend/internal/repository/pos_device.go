package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PosDeviceRepository interface {
	UpsertPendingByToken(ctx context.Context, name, model, os, token string) (*domain.PosDevice, error)
	FindByToken(ctx context.Context, token string) (*domain.PosDevice, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.PosDevice, error)
	List(ctx context.Context, status *domain.PosRequestStatus) ([]*domain.PosDevice, error)
	UpdateConfig(ctx context.Context, id bson.ObjectID, cardCapable *bool, printerMAC *string, printerUUID *string) error
	UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.PosRequestStatus, decidedBy *bson.ObjectID, decidedAt time.Time) error
}

type posDeviceRepository struct {
	collection *mongo.Collection
}

func NewPosDeviceRepository(db *database.MongoDB) PosDeviceRepository {
	return &posDeviceRepository{collection: db.Database.Collection(database.PosDevicesCollection)}
}

func (r *posDeviceRepository) UpsertPendingByToken(ctx context.Context, name, model, os, token string) (*domain.PosDevice, error) {
	now := time.Now().UTC()
	existing, err := r.FindByToken(ctx, token)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	if err == nil && existing != nil {
		status := existing.Status
		if status == "" {
			if existing.Approved {
				status = domain.PosRequestStatusApproved
			} else {
				status = domain.PosRequestStatusPending
			}
		}
		if status != domain.PosRequestStatusApproved {
			status = domain.PosRequestStatusPending
		}
		set := bson.M{
			"name":       name,
			"model":      model,
			"os":         os,
			"status":     status,
			"approved":   status == domain.PosRequestStatusApproved,
			"updated_at": now,
		}
		update := bson.M{"$set": set}
		if status != domain.PosRequestStatusApproved {
			update["$unset"] = bson.M{"approved_at": ""}
		}
		if _, err := r.collection.UpdateByID(ctx, existing.ID, update); err != nil {
			return nil, err
		}
		return r.FindByID(ctx, existing.ID)
	}
	// otherwise create fresh pending device
	dev := &domain.PosDevice{
		ID:          bson.NewObjectID(),
		Name:        name,
		Model:       model,
		OS:          os,
		DeviceToken: token,
		Status:      domain.PosRequestStatusPending,
		Approved:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := r.collection.InsertOne(ctx, dev); err != nil {
		return nil, err
	}
	return dev, nil
}

func (r *posDeviceRepository) FindByToken(ctx context.Context, token string) (*domain.PosDevice, error) {
	var d domain.PosDevice
	if err := r.collection.FindOne(ctx, bson.M{"device_token": token}).Decode(&d); err != nil {
		return nil, err
	}
	return normalizePosDevice(&d), nil
}

func (r *posDeviceRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.PosDevice, error) {
	var d domain.PosDevice
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&d); err != nil {
		return nil, err
	}
	return normalizePosDevice(&d), nil
}

func (r *posDeviceRepository) List(ctx context.Context, status *domain.PosRequestStatus) ([]*domain.PosDevice, error) {
	filter := bson.M{}
	if status != nil && *status != "" {
		clauses := []bson.M{{"status": *status}}
		if *status == domain.PosRequestStatusApproved {
			clauses = append(clauses, bson.M{"status": bson.M{"$exists": false}, "approved": true})
		}
		if *status == domain.PosRequestStatusPending {
			clauses = append(clauses, bson.M{"status": bson.M{"$exists": false}, "approved": bson.M{"$ne": true}})
		}
		filter["$or"] = clauses
	}
	cur, err := r.collection.Find(ctx, filter)
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
		out = append(out, normalizePosDevice(&it))
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

func (r *posDeviceRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.PosRequestStatus, decidedBy *bson.ObjectID, decidedAt time.Time) error {
	set := bson.M{
		"status":     status,
		"approved":   status == domain.PosRequestStatusApproved,
		"updated_at": time.Now().UTC(),
		"decided_at": decidedAt,
	}
	unset := bson.M{}
	if decidedBy != nil {
		set["decided_by"] = *decidedBy
	} else {
		unset["decided_by"] = ""
	}
	if status == domain.PosRequestStatusApproved {
		set["approved_at"] = decidedAt
	} else {
		unset["approved_at"] = ""
	}
	update := bson.M{"$set": set}
	if len(unset) > 0 {
		update["$unset"] = unset
	}
	_, err := r.collection.UpdateByID(ctx, id, update)
	return err
}

func normalizePosDevice(dev *domain.PosDevice) *domain.PosDevice {
	if dev == nil {
		return dev
	}
	if dev.Status == "" {
		if dev.Approved {
			dev.Status = domain.PosRequestStatusApproved
		} else {
			dev.Status = domain.PosRequestStatusPending
		}
	}
	dev.Approved = dev.Status == domain.PosRequestStatusApproved
	return dev
}
