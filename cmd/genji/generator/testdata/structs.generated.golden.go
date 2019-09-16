// Code generated by genji.
// DO NOT EDIT!

package testdata

import (
	"errors"

	"github.com/asdine/genji/field"
	"github.com/asdine/genji/record"
	"github.com/asdine/genji/value"
)

// GetField implements the field method of the record.Record interface.
func (b *Basic) GetField(name string) (field.Field, error) {
	switch name {
	case "A":
		return field.NewString("A", b.A), nil
	case "B":
		return field.NewInt("B", b.B), nil
	case "C":
		return field.NewInt32("C", b.C), nil
	case "D":
		return field.NewInt32("D", b.D), nil
	}

	return field.Field{}, errors.New("unknown field")
}

// Iterate through all the fields one by one and pass each of them to the given function.
// It the given function returns an error, the iteration is interrupted.
func (b *Basic) Iterate(fn func(field.Field) error) error {
	var err error

	err = fn(field.NewString("A", b.A))
	if err != nil {
		return err
	}

	err = fn(field.NewInt("B", b.B))
	if err != nil {
		return err
	}

	err = fn(field.NewInt32("C", b.C))
	if err != nil {
		return err
	}

	err = fn(field.NewInt32("D", b.D))
	if err != nil {
		return err
	}

	return nil
}

// ScanRecord extracts fields from record and assigns them to the struct fields.
// It implements the record.Scanner interface.
func (b *Basic) ScanRecord(rec record.Record) error {
	return rec.Iterate(func(f field.Field) error {
		var err error

		switch f.Name {
		case "A":
			b.A, err = value.DecodeString(f.Data)
		case "B":
			b.B, err = value.DecodeInt(f.Data)
		case "C":
			b.C, err = value.DecodeInt32(f.Data)
		case "D":
			b.D, err = value.DecodeInt32(f.Data)
		}
		return err
	})
}

// Scan extracts fields from src and assigns them to the struct fields.
// It implements the driver.Scanner interface.
func (b *Basic) Scan(src interface{}) error {
	r, ok := src.(record.Record)
	if !ok {
		return errors.New("unable to scan record from src")
	}

	return b.ScanRecord(r)
}

// GetField implements the field method of the record.Record interface.
func (b *basic) GetField(name string) (field.Field, error) {
	switch name {
	case "A":
		return field.NewBytes("A", b.A), nil
	case "B":
		return field.NewUint16("B", b.B), nil
	case "C":
		return field.NewFloat32("C", b.C), nil
	case "D":
		return field.NewFloat32("D", b.D), nil
	}

	return field.Field{}, errors.New("unknown field")
}

// Iterate through all the fields one by one and pass each of them to the given function.
// It the given function returns an error, the iteration is interrupted.
func (b *basic) Iterate(fn func(field.Field) error) error {
	var err error

	err = fn(field.NewBytes("A", b.A))
	if err != nil {
		return err
	}

	err = fn(field.NewUint16("B", b.B))
	if err != nil {
		return err
	}

	err = fn(field.NewFloat32("C", b.C))
	if err != nil {
		return err
	}

	err = fn(field.NewFloat32("D", b.D))
	if err != nil {
		return err
	}

	return nil
}

// ScanRecord extracts fields from record and assigns them to the struct fields.
// It implements the record.Scanner interface.
func (b *basic) ScanRecord(rec record.Record) error {
	return rec.Iterate(func(f field.Field) error {
		var err error

		switch f.Name {
		case "A":
			b.A, err = value.DecodeBytes(f.Data)
		case "B":
			b.B, err = value.DecodeUint16(f.Data)
		case "C":
			b.C, err = value.DecodeFloat32(f.Data)
		case "D":
			b.D, err = value.DecodeFloat32(f.Data)
		}
		return err
	})
}

// Scan extracts fields from src and assigns them to the struct fields.
// It implements the driver.Scanner interface.
func (b *basic) Scan(src interface{}) error {
	r, ok := src.(record.Record)
	if !ok {
		return errors.New("unable to scan record from src")
	}

	return b.ScanRecord(r)
}

// GetField implements the field method of the record.Record interface.
func (p *Pk) GetField(name string) (field.Field, error) {
	switch name {
	case "A":
		return field.NewString("A", p.A), nil
	case "B":
		return field.NewInt64("B", p.B), nil
	}

	return field.Field{}, errors.New("unknown field")
}

// Iterate through all the fields one by one and pass each of them to the given function.
// It the given function returns an error, the iteration is interrupted.
func (p *Pk) Iterate(fn func(field.Field) error) error {
	var err error

	err = fn(field.NewString("A", p.A))
	if err != nil {
		return err
	}

	err = fn(field.NewInt64("B", p.B))
	if err != nil {
		return err
	}

	return nil
}

// ScanRecord extracts fields from record and assigns them to the struct fields.
// It implements the record.Scanner interface.
func (p *Pk) ScanRecord(rec record.Record) error {
	return rec.Iterate(func(f field.Field) error {
		var err error

		switch f.Name {
		case "A":
			p.A, err = value.DecodeString(f.Data)
		case "B":
			p.B, err = value.DecodeInt64(f.Data)
		}
		return err
	})
}

// Scan extracts fields from src and assigns them to the struct fields.
// It implements the driver.Scanner interface.
func (p *Pk) Scan(src interface{}) error {
	r, ok := src.(record.Record)
	if !ok {
		return errors.New("unable to scan record from src")
	}

	return p.ScanRecord(r)
}

// PrimaryKey returns the primary key. It implements the table.PrimaryKeyer interface.
func (p *Pk) PrimaryKey() ([]byte, error) {
	return value.EncodeInt64(p.B), nil
}
