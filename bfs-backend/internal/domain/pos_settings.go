package domain

import "time"

type PosFulfillmentMode string

const (
	PosModeQRCode PosFulfillmentMode = "QR_CODE"
	PosModeJeton  PosFulfillmentMode = "JETON"
)

type PosSettings struct {
	ID        string             `bson:"_id" json:"id"`
	Mode      PosFulfillmentMode `bson:"mode" json:"mode"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
}
