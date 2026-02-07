package api

import (
	"backend/internal/blobstore"
	"backend/internal/generated/api/generated"
	"backend/internal/inventory"
	"backend/internal/repository"
	"backend/internal/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Handlers implements generated.ServerInterface by injecting services directly.
type Handlers struct {
	generated.Unimplemented

	categories   service.CategoryService
	products     service.ProductService
	orders       service.OrderService
	payments     service.PaymentService
	pos          service.POSService
	settings     service.SettingsService
	stations     service.StationService
	invites      service.AdminInviteService
	email        service.EmailService
	users        service.UserService
	devices      service.DeviceService
	club100      service.Club100Service
	verification repository.VerificationRepository
	idempotency  repository.IdempotencyRepository
	blobStore    *blobstore.Client
	inventoryHub *inventory.Hub

	logger *zap.Logger
}

// Compile-time check that Handlers implements ServerInterface.
var _ generated.ServerInterface = (*Handlers)(nil)

type HandlersDeps struct {
	fx.In

	Categories   service.CategoryService
	Products     service.ProductService
	Orders       service.OrderService
	Payments     service.PaymentService
	POS          service.POSService
	Settings     service.SettingsService
	Stations     service.StationService
	Invites      service.AdminInviteService
	Email        service.EmailService
	Users        service.UserService
	Devices      service.DeviceService
	Club100      service.Club100Service
	Verification repository.VerificationRepository
	Idempotency  repository.IdempotencyRepository
	BlobStore    *blobstore.Client `optional:"true"`
	InventoryHub *inventory.Hub
	Logger       *zap.Logger
}

// NewHandlers creates a new Handlers with all required service dependencies.
func NewHandlers(deps HandlersDeps) *Handlers {
	return &Handlers{
		categories:   deps.Categories,
		products:     deps.Products,
		orders:       deps.Orders,
		payments:     deps.Payments,
		pos:          deps.POS,
		settings:     deps.Settings,
		stations:     deps.Stations,
		invites:      deps.Invites,
		email:        deps.Email,
		users:        deps.Users,
		devices:      deps.Devices,
		club100:      deps.Club100,
		verification: deps.Verification,
		idempotency:  deps.Idempotency,
		blobStore:    deps.BlobStore,
		inventoryHub: deps.InventoryHub,
		logger:       deps.Logger,
	}
}
