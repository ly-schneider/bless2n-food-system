package service

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type StationService interface {
	RequestVerification(ctx context.Context, name, model, os, deviceKey string) (*domain.Station, error)
	GetStationByKey(ctx context.Context, deviceKey string) (*domain.Station, error)
	VerifyQR(ctx context.Context, code string) (bson.ObjectID, error)
	AssignedItemsForOrder(ctx context.Context, stationID, orderID bson.ObjectID) ([]*domain.OrderItem, error)
	RedeemAssigned(ctx context.Context, stationID bson.ObjectID, orderID bson.ObjectID, idemKey string) (map[string]any, error)
	MakePickupQR(ctx context.Context, orderID bson.ObjectID) (string, error)
}

type stationService struct {
	cfg          config.Config
	stations     repository.StationRepository
	stationProds repository.StationProductRepository
	orders       repository.OrderRepository
	orderItems   repository.OrderItemRepository
	idempotency  repository.IdempotencyRepository
}

func NewStationService(
	cfg config.Config,
	stations repository.StationRepository,
	stationProds repository.StationProductRepository,
	orders repository.OrderRepository,
	orderItems repository.OrderItemRepository,
	idempotency repository.IdempotencyRepository,
) StationService {
	return &stationService{cfg: cfg, stations: stations, stationProds: stationProds, orders: orders, orderItems: orderItems, idempotency: idempotency}
}

func (s *stationService) RequestVerification(ctx context.Context, name, model, os, deviceKey string) (*domain.Station, error) {
	if name == "" || deviceKey == "" {
		return nil, errors.New("invalid_payload")
	}
	st, err := s.stations.UpsertPendingByDeviceKey(ctx, name, model, os, deviceKey)
	if err != nil {
		return nil, err
	}
	return st, nil
}

func (s *stationService) GetStationByKey(ctx context.Context, deviceKey string) (*domain.Station, error) {
	return s.stations.FindByDeviceKey(ctx, deviceKey)
}

// VerifyQR parses a QR code string and validates HMAC signature and freshness.
// Expected minimal format: v=1;orderId=<hex>;issuedAt=<rfc3339>;sig=<base64url>
func (s *stationService) VerifyQR(ctx context.Context, code string) (bson.ObjectID, error) { // nolint: revive
	parts := parseKV(code)
	if parts["v"] != "1" {
		return bson.NilObjectID, errors.New("invalid_qr_version")
	}
	oid, err := bson.ObjectIDFromHex(parts["orderId"])
	if err != nil {
		return bson.NilObjectID, errors.New("invalid_order_id")
	}
	issuedAtStr := parts["issuedAt"]
	if issuedAtStr == "" {
		return bson.NilObjectID, errors.New("invalid_issued_at")
	}
	ia, err := time.Parse(time.RFC3339, issuedAtStr)
	if err != nil {
		return bson.NilObjectID, errors.New("invalid_issued_at")
	}
	maxAge := time.Duration(s.cfg.Stations.QRMaxAgeSeconds) * time.Second
	if maxAge <= 0 {
		maxAge = 10 * time.Minute
	}
	if time.Since(ia) > maxAge {
		return bson.NilObjectID, errors.New("qr_expired")
	}
	// Verify signature
	sig := parts["sig"]
	if sig == "" {
		return bson.NilObjectID, errors.New("missing_sig")
	}
	payload := fmt.Sprintf("v=%s;orderId=%s;issuedAt=%s", parts["v"], parts["orderId"], issuedAtStr)
	if !verifyHMAC(payload, sig, s.cfg.Stations.QRSecret) {
		return bson.NilObjectID, errors.New("bad_sig")
	}
	return oid, nil
}

func (s *stationService) AssignedItemsForOrder(ctx context.Context, stationID, orderID bson.ObjectID) ([]*domain.OrderItem, error) {
	// filter order items to only product IDs assigned to this station
	pids, err := s.stationProds.ListProductIDsByStation(ctx, stationID)
	if err != nil {
		return nil, err
	}
	if len(pids) == 0 {
		return []*domain.OrderItem{}, nil
	}
	// Inline query here to avoid expanding repository surface
	// (Using orderItems collection via aggregation-like filter)
	return s.orderItems.FindByFilter(ctx, bson.M{"order_id": orderID, "product_id": bson.M{"$in": pids}})
}

func (s *stationService) RedeemAssigned(ctx context.Context, stationID bson.ObjectID, orderID bson.ObjectID, idemKey string) (map[string]any, error) {
	scope := fmt.Sprintf("station:%s:order:%s", stationID.Hex(), orderID.Hex())
	if idemKey != "" {
		if rec, err := s.idempotency.Get(ctx, scope, idemKey); err == nil && rec != nil {
			return rec.Response, nil
		}
	}
	pids, err := s.stationProds.ListProductIDsByStation(ctx, stationID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	// Perform single conditional write to mark items as redeemed
	// Update filter: this order, assigned products, not yet redeemed
	// We need direct collection access; extend repository with method via type assertion
	matched, modified, err := s.orderItems.UpdateRedeemForOrderByProductIDs(ctx, orderID, pids, now)
	if err != nil {
		return nil, err
	}
	// Load current assigned items state for receipt
	assigned, err := s.AssignedItemsForOrder(ctx, stationID, orderID)
	if err != nil {
		return nil, err
	}
	// Build response
	resp := map[string]any{
		"orderId":    orderID.Hex(),
		"stationId":  stationID.Hex(),
		"matched":    matched,
		"redeemed":   modified,
		"items":      toPublicOrderItems(assigned),
		"redeemedAt": now,
	}
	// If modified == 0, handler may choose to return HTTP 409 Conflict.
	if idemKey != "" {
		_, _ = s.idempotency.SaveIfAbsent(ctx, &domain.IdempotencyRecord{Key: idemKey, Scope: scope, Response: resp, CreatedAt: now})
	}
	return resp, nil
}

// MakePickupQR generates a minimal signed QR payload for an orderId
// Format: v=1;orderId=<hex>;issuedAt=<RFC3339>;sig=<base64url(HMAC256)>
func (s *stationService) MakePickupQR(ctx context.Context, orderID bson.ObjectID) (string, error) {
	if orderID.IsZero() {
		return "", errors.New("invalid_order_id")
	}
	ia := time.Now().UTC().Format(time.RFC3339)
	payload := fmt.Sprintf("v=1;orderId=%s;issuedAt=%s", orderID.Hex(), ia)
	sig := signHMAC(payload, s.cfg.Stations.QRSecret)
	return payload + ";sig=" + sig, nil
}

func parseKV(s string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(s, ";") {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v, _ := url.QueryUnescape(strings.TrimSpace(kv[1]))
		out[k] = v
	}
	return out
}

func verifyHMAC(payload, sigB64url, secret string) bool {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(payload))
	mac := m.Sum(nil)
	dec, err := base64.RawURLEncoding.DecodeString(sigB64url)
	if err != nil {
		return false
	}
	return hmac.Equal(mac, dec)
}

func signHMAC(payload, secret string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(payload))
	mac := m.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(mac)
}

func toPublicOrderItems(items []*domain.OrderItem) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		var parentID *string
		if it.ParentItemID != nil {
			s := it.ParentItemID.Hex()
			parentID = &s
		}
		var msID *string
		if it.MenuSlotID != nil {
			s := it.MenuSlotID.Hex()
			msID = &s
		}
		out = append(out, map[string]any{
			"id":           it.ID.Hex(),
			"orderId":      it.OrderID.Hex(),
			"productId":    it.ProductID.Hex(),
			"title":        it.Title,
			"quantity":     it.Quantity,
			"isRedeemed":   it.IsRedeemed,
			"parentItemId": parentID,
			"menuSlotId":   msID,
			"menuSlotName": it.MenuSlotName,
		})
	}
	return out
}
