package model

import (
	"database/sql/driver"
	"fmt"
)

// UserRole represents user roles.
type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleCustomer UserRole = "customer"
)

func (r UserRole) Value() (driver.Value, error) { return string(r), nil }
func (r *UserRole) Scan(value any) error {
	if v, ok := value.(string); ok {
		*r = UserRole(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into UserRole", value)
}

// OrderStatus represents order states.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
)

func (s OrderStatus) Value() (driver.Value, error) { return string(s), nil }
func (s *OrderStatus) Scan(value any) error {
	if v, ok := value.(string); ok {
		*s = OrderStatus(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into OrderStatus", value)
}

// OrderOrigin represents where the order originated.
type OrderOrigin string

const (
	OrderOriginShop OrderOrigin = "shop"
	OrderOriginPOS  OrderOrigin = "pos"
)

func (o OrderOrigin) Value() (driver.Value, error) { return string(o), nil }
func (o *OrderOrigin) Scan(value any) error {
	if v, ok := value.(string); ok {
		*o = OrderOrigin(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into OrderOrigin", value)
}

// ProductType represents product types.
type ProductType string

const (
	ProductTypeSimple ProductType = "simple"
	ProductTypeMenu   ProductType = "menu"
)

func (t ProductType) Value() (driver.Value, error) { return string(t), nil }
func (t *ProductType) Scan(value any) error {
	if v, ok := value.(string); ok {
		*t = ProductType(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into ProductType", value)
}

// OrderItemType represents order item types.
type OrderItemType string

const (
	OrderItemTypeSimple    OrderItemType = "simple"
	OrderItemTypeBundle    OrderItemType = "bundle"
	OrderItemTypeComponent OrderItemType = "component"
)

func (t OrderItemType) Value() (driver.Value, error) { return string(t), nil }
func (t *OrderItemType) Scan(value any) error {
	if v, ok := value.(string); ok {
		*t = OrderItemType(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into OrderItemType", value)
}

// InventoryReason represents inventory adjustment reasons.
type InventoryReason string

const (
	InventoryReasonOpeningBalance InventoryReason = "opening_balance"
	InventoryReasonSale           InventoryReason = "sale"
	InventoryReasonRefund         InventoryReason = "refund"
	InventoryReasonManualAdjust   InventoryReason = "manual_adjust"
	InventoryReasonCorrection     InventoryReason = "correction"
)

func (r InventoryReason) Value() (driver.Value, error) { return string(r), nil }
func (r *InventoryReason) Scan(value any) error {
	if v, ok := value.(string); ok {
		*r = InventoryReason(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into InventoryReason", value)
}

// CommonStatus represents common approval states.
type CommonStatus string

const (
	CommonStatusPending  CommonStatus = "pending"
	CommonStatusApproved CommonStatus = "approved"
	CommonStatusRejected CommonStatus = "rejected"
	CommonStatusRevoked  CommonStatus = "revoked"
)

func (s CommonStatus) Value() (driver.Value, error) { return string(s), nil }
func (s *CommonStatus) Scan(value any) error {
	if v, ok := value.(string); ok {
		*s = CommonStatus(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into CommonStatus", value)
}

// PosFulfillmentMode represents POS fulfillment modes.
type PosFulfillmentMode string

const (
	PosFulfillmentModeQRCode PosFulfillmentMode = "QR_CODE"
	PosFulfillmentModeJeton  PosFulfillmentMode = "JETON"
)

func (m PosFulfillmentMode) Value() (driver.Value, error) { return string(m), nil }
func (m *PosFulfillmentMode) Scan(value any) error {
	if v, ok := value.(string); ok {
		*m = PosFulfillmentMode(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into PosFulfillmentMode", value)
}

// PaymentMethod represents payment methods.
type PaymentMethod string

const (
	PaymentMethodCash  PaymentMethod = "CASH"
	PaymentMethodCard  PaymentMethod = "CARD"
	PaymentMethodTwint PaymentMethod = "TWINT"
)

func (m PaymentMethod) Value() (driver.Value, error) { return string(m), nil }
func (m *PaymentMethod) Scan(value any) error {
	if v, ok := value.(string); ok {
		*m = PaymentMethod(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into PaymentMethod", value)
}

// DeviceType represents device types.
type DeviceType string

const (
	DeviceTypePOS     DeviceType = "POS"
	DeviceTypeStation DeviceType = "STATION"
)

func (t DeviceType) Value() (driver.Value, error) { return string(t), nil }
func (t *DeviceType) Scan(value any) error {
	if v, ok := value.(string); ok {
		*t = DeviceType(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into DeviceType", value)
}
