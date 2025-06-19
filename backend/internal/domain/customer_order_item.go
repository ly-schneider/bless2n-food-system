package domain

import (
	"backend/internal/model"
)

type CustomerOrderItem struct {
	CustomerOrderID model.NanoID14 `gorm:"type:nano_id;primaryKey"                                           json:"customer_order_id"`
	EventProductID  model.NanoID14 `gorm:"type:nano_id;primaryKey"                                           json:"event_product_id"`
	Quantity        int            `gorm:"not null;check:quantity > 0"                                       json:"quantity"         validate:"required,gt=0"`
	PricePerUnit    float64        `gorm:"type:numeric(6,2);not null;check:price_per_unit >= 0"              json:"price_per_unit"   validate:"required"`
	CustomerOrder   *CustomerOrder `gorm:"constraint:OnDelete:CASCADE"                                       json:"customer_order,omitempty"`
	EventProduct    *EventProduct  `gorm:"constraint:OnDelete:RESTRICT"                                       json:"event_product,omitempty"`
}
