package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"backend/internal/id"
)

// nanoidPK is the shared primary-key field. The column is varchar(36), not
// varchar(12), so it can also hold the legacy uuidv7 ids that pre-date the
// migration.
func nanoidPK() ent.Field {
	return field.String("id").
		MaxLen(36).
		NotEmpty().
		DefaultFunc(id.New).
		Immutable()
}
