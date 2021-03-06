// Code generated by entc, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/Jameslikestea/grm/internal/storage/ent/reference"
)

// Reference is the model entity for the Reference schema.
type Reference struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Package holds the value of the "package" field.
	Package string `json:"package,omitempty"`
	// Ref holds the value of the "ref" field.
	Ref string `json:"ref,omitempty"`
	// Hash holds the value of the "hash" field.
	Hash string `json:"hash,omitempty"`
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Reference) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case reference.FieldID:
			values[i] = new(sql.NullInt64)
		case reference.FieldPackage, reference.FieldRef, reference.FieldHash:
			values[i] = new(sql.NullString)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Reference", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Reference fields.
func (r *Reference) assignValues(columns []string, values []interface{}) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case reference.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			r.ID = int(value.Int64)
		case reference.FieldPackage:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field package", values[i])
			} else if value.Valid {
				r.Package = value.String
			}
		case reference.FieldRef:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field ref", values[i])
			} else if value.Valid {
				r.Ref = value.String
			}
		case reference.FieldHash:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field hash", values[i])
			} else if value.Valid {
				r.Hash = value.String
			}
		}
	}
	return nil
}

// Update returns a builder for updating this Reference.
// Note that you need to call Reference.Unwrap() before calling this method if this Reference
// was returned from a transaction, and the transaction was committed or rolled back.
func (r *Reference) Update() *ReferenceUpdateOne {
	return (&ReferenceClient{config: r.config}).UpdateOne(r)
}

// Unwrap unwraps the Reference entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (r *Reference) Unwrap() *Reference {
	tx, ok := r.config.driver.(*txDriver)
	if !ok {
		panic("ent: Reference is not a transactional entity")
	}
	r.config.driver = tx.drv
	return r
}

// String implements the fmt.Stringer.
func (r *Reference) String() string {
	var builder strings.Builder
	builder.WriteString("Reference(")
	builder.WriteString(fmt.Sprintf("id=%v", r.ID))
	builder.WriteString(", package=")
	builder.WriteString(r.Package)
	builder.WriteString(", ref=")
	builder.WriteString(r.Ref)
	builder.WriteString(", hash=")
	builder.WriteString(r.Hash)
	builder.WriteByte(')')
	return builder.String()
}

// References is a parsable slice of Reference.
type References []*Reference

func (r References) config(cfg config) {
	for _i := range r {
		r[_i].config = cfg
	}
}
