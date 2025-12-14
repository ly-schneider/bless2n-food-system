package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type StationRepository interface {
	UpsertPendingByDeviceKey(ctx context.Context, name, model, os, deviceKey string) (*domain.Station, error)
	FindByDeviceKey(ctx context.Context, deviceKey string) (*domain.Station, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.Station, error)
	List(ctx context.Context, status *domain.StationRequestStatus) ([]*domain.Station, error)
	UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.StationRequestStatus, decidedBy *bson.ObjectID, decidedAt time.Time) error
}

type stationRepository struct {
	collection *mongo.Collection
}

func NewStationRepository(db *database.MongoDB) StationRepository {
	return &stationRepository{
		collection: db.Database.Collection(database.StationsCollection),
	}
}

func (r *stationRepository) UpsertPendingByDeviceKey(ctx context.Context, name, model, os, deviceKey string) (*domain.Station, error) {
	now := time.Now().UTC()
	// If the station already exists, bump metadata and reset status if it is not approved
	existing, err := r.FindByDeviceKey(ctx, deviceKey)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	if err == nil && existing != nil {
		status := existing.Status
		if status == "" {
			if existing.Approved {
				status = domain.StationRequestStatusApproved
			} else {
				status = domain.StationRequestStatusPending
			}
		}
		if status != domain.StationRequestStatusApproved {
			status = domain.StationRequestStatusPending
		}
		set := bson.M{
			"name":       name,
			"model":      model,
			"os":         os,
			"status":     status,
			"approved":   status == domain.StationRequestStatusApproved,
			"updated_at": now,
			"expires_at": now.Add(30 * 24 * time.Hour),
		}
		update := bson.M{"$set": set}
		if status != domain.StationRequestStatusApproved {
			update["$unset"] = bson.M{"approved_at": ""}
		}
		if _, err := r.collection.UpdateByID(ctx, existing.ID, update); err != nil {
			return nil, err
		}
		return r.FindByID(ctx, existing.ID)
	}
	// Otherwise insert a fresh pending station record
	expiresAt := now.Add(30 * 24 * time.Hour)
	st := &domain.Station{
		ID:        bson.NewObjectID(),
		Name:      name,
		Model:     model,
		OS:        os,
		DeviceKey: deviceKey,
		Status:    domain.StationRequestStatusPending,
		Approved:  false,
		ExpiresAt: &expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := r.collection.InsertOne(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

func (r *stationRepository) FindByDeviceKey(ctx context.Context, deviceKey string) (*domain.Station, error) {
	var st domain.Station
	if err := r.collection.FindOne(ctx, bson.M{"device_key": deviceKey}).Decode(&st); err != nil {
		return nil, err
	}
	return normalizeStation(&st), nil
}

func (r *stationRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Station, error) {
	var st domain.Station
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&st); err != nil {
		return nil, err
	}
	return normalizeStation(&st), nil
}

func (r *stationRepository) List(ctx context.Context, status *domain.StationRequestStatus) ([]*domain.Station, error) {
	filter := bson.M{}
	if status != nil && *status != "" {
		clauses := []bson.M{{"status": *status}}
		if *status == domain.StationRequestStatusApproved {
			clauses = append(clauses, bson.M{"status": bson.M{"$exists": false}, "approved": true})
		}
		if *status == domain.StationRequestStatusPending {
			clauses = append(clauses, bson.M{"status": bson.M{"$exists": false}, "approved": bson.M{"$ne": true}})
		}
		filter["$or"] = clauses
	}
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []*domain.Station
	for cur.Next(ctx) {
		var st domain.Station
		if err := cur.Decode(&st); err != nil {
			return nil, err
		}
		out = append(out, normalizeStation(&st))
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *stationRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.StationRequestStatus, decidedBy *bson.ObjectID, decidedAt time.Time) error {
	set := bson.M{
		"status":     status,
		"approved":   status == domain.StationRequestStatusApproved,
		"updated_at": time.Now().UTC(),
		"decided_at": decidedAt,
	}
	unset := bson.M{}
	if decidedBy != nil {
		set["decided_by"] = *decidedBy
	} else {
		unset["decided_by"] = ""
	}
	if status == domain.StationRequestStatusApproved {
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

func normalizeStation(st *domain.Station) *domain.Station {
	if st == nil {
		return st
	}
	if st.Status == "" {
		if st.Approved {
			st.Status = domain.StationRequestStatusApproved
		} else {
			st.Status = domain.StationRequestStatusPending
		}
	}
	st.Approved = st.Status == domain.StationRequestStatusApproved
	return st
}
