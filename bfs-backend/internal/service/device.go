package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/device"
	"backend/internal/generated/ent/devicebinding"
	"backend/internal/repository"

	"github.com/google/uuid"
)

// PairingResult is the result of creating a new device pairing code.
type PairingResult struct {
	Code      string
	ExpiresAt time.Time
}

// PairingStatusResult is the result of polling a device pairing status.
type PairingStatusResult struct {
	Status string
	Token  *string
	Device *ent.Device
}

type DeviceService interface {
	ListAll(ctx context.Context, deviceType *string, status *string) ([]*ent.Device, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Device, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	CreatePairing(ctx context.Context, deviceKey, name, deviceType string, model, os *string) (*PairingResult, error)
	GetPairingStatus(ctx context.Context, code string) (*PairingStatusResult, error)
	CompletePairing(ctx context.Context, code string, adminUserID string) (*ent.Device, error)
}

type deviceService struct {
	client      *ent.Client
	devices     repository.DeviceRepository
	bindings    repository.DeviceBindingRepository
	sessionRepo repository.SessionRepository
}

func NewDeviceService(
	client *ent.Client,
	devices repository.DeviceRepository,
	bindings repository.DeviceBindingRepository,
	sessionRepo repository.SessionRepository,
) DeviceService {
	return &deviceService{
		client:      client,
		devices:     devices,
		bindings:    bindings,
		sessionRepo: sessionRepo,
	}
}

func (s *deviceService) ListAll(ctx context.Context, deviceType *string, status *string) ([]*ent.Device, error) {
	q := s.client.Device.Query()
	if deviceType != nil && *deviceType != "" {
		q = q.Where(device.TypeEQ(device.Type(*deviceType)))
	}
	if status != nil && *status != "" {
		q = q.Where(device.StatusEQ(device.Status(*status)))
	}
	return q.Order(device.ByCreatedAt()).All(ctx)
}

func (s *deviceService) GetByID(ctx context.Context, id uuid.UUID) (*ent.Device, error) {
	return s.devices.GetByID(ctx, id)
}

func (s *deviceService) Revoke(ctx context.Context, id uuid.UUID) error {
	d, err := s.devices.GetByID(ctx, id)
	if err != nil {
		return err
	}
	_, err = s.devices.Update(ctx, d.ID, d.Name, d.DeviceKey, d.Type, device.StatusRejected,
		d.Model, d.Os, d.DecidedBy, d.DecidedAt, d.ExpiresAt,
		d.PendingSessionToken, d.PairingCode, d.PairingCodeExpiresAt)
	return err
}

func (s *deviceService) CreatePairing(ctx context.Context, deviceKey, name, deviceType string, model, os *string) (*PairingResult, error) {
	mdl := ""
	if model != nil {
		mdl = *model
	}
	osStr := ""
	if os != nil {
		osStr = *os
	}

	d, err := s.devices.GeneratePairingCode(ctx, name, mdl, osStr, deviceKey, device.Type(deviceType))
	if err != nil {
		return nil, err
	}

	if d.PairingCode == nil || d.PairingCodeExpiresAt == nil {
		return nil, errors.New("pairing code generation failed")
	}

	return &PairingResult{
		Code:      *d.PairingCode,
		ExpiresAt: *d.PairingCodeExpiresAt,
	}, nil
}

func (s *deviceService) GetPairingStatus(ctx context.Context, code string) (*PairingStatusResult, error) {
	d, err := s.devices.GetByPairingCode(ctx, code)
	if err != nil {
		return nil, err
	}

	result := &PairingStatusResult{
		Status: string(d.Status),
		Device: d,
	}

	// If device has been approved and has a pending session token, return it.
	if d.Status == device.StatusApproved && d.PendingSessionToken != nil {
		result.Token = d.PendingSessionToken
	}

	return result, nil
}

// deviceSessionExpiry matches the Better Auth session TTL (90 days).
const deviceSessionExpiry = 90 * 24 * time.Hour

func (s *deviceService) CompletePairing(ctx context.Context, code string, adminUserID string) (*ent.Device, error) {
	d, err := s.devices.GetByPairingCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if d.Status != device.StatusPending {
		return nil, fmt.Errorf("device is not pending approval")
	}

	// Create a Better Auth session for the admin user. The session token
	// is what the device will send as a Bearer token. The device auth
	// middleware validates it against both the device_binding (hash) and
	// the session table.
	sessionToken, err := s.sessionRepo.CreateSession(ctx, adminUserID, deviceSessionExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to create device session: %w", err)
	}

	tokenHash := repository.HashToken(sessionToken)

	// Approve the device.
	now := time.Now().UTC()
	updated, err := s.devices.Update(ctx, d.ID, d.Name, d.DeviceKey, d.Type, device.StatusApproved,
		d.Model, d.Os, &adminUserID, &now, d.ExpiresAt,
		nil, d.PairingCode, d.PairingCodeExpiresAt)
	if err != nil {
		return nil, err
	}

	// Create a device_binding row so the middleware can find it by token hash.
	bindingType := devicebinding.DeviceTypePOS
	if d.Type == device.TypeSTATION {
		bindingType = devicebinding.DeviceTypeSTATION
	}
	deviceID := d.ID
	if _, err := s.bindings.Create(ctx, bindingType, tokenHash, adminUserID, &d.Name, &deviceID, nil); err != nil {
		return nil, fmt.Errorf("failed to create device binding: %w", err)
	}

	// Store the raw session token so the polling device can pick it up.
	if err := s.devices.SetPendingSessionToken(ctx, d.ID, sessionToken); err != nil {
		return nil, err
	}

	return updated, nil
}
