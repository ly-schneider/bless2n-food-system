package model

import (
	"backend/internal/utils"
	"database/sql/driver"
	"errors"
)

// NanoID14 is a 14-character public ID stored as CHAR(14) in PostgreSQL.
type NanoID14 string

/* ---------- driver.Valuer ---------- */

func (n NanoID14) Value() (driver.Value, error) {
	if err := utils.Validate("id", string(n)); err != nil {
		return nil, err
	}
	return string(n), nil
}

/* ---------- sql.Scanner ---------- */

func (n *NanoID14) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*n = NanoID14(v)
	case []byte:
		*n = NanoID14(v)
	default:
		return errors.New("NanoID14: unsupported source type")
	}
	return utils.Validate("id", string(*n))
}
