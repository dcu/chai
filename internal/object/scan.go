package object

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/chaisql/chai/internal/types"
	"github.com/cockroachdb/errors"
)

// A Scanner can iterate over an object and scan all the fields.
type Scanner interface {
	ScanObject(types.Object) error
}

// Scan each field of the object into the given variables.
func Scan(d types.Object, targets ...interface{}) error {
	var i int

	return d.Iterate(func(f string, v types.Value) error {
		if i >= len(targets) {
			return errors.New("target list too small")
		}

		target := targets[i]
		i++

		ref, ok := target.(reflect.Value)
		if !ok {
			ref = reflect.ValueOf(target)
			if !ref.IsValid() {
				return &ErrUnsupportedType{target, fmt.Sprintf("Parameter %d is not valid", i)}
			}
		}

		return scanValue(v, ref)
	})
}

// StructScan scans d into t. t is expected to be a pointer to a struct.
//
// By default, each struct field name is lowercased and the object's GetByField method
// is called with that name. If there is a match, the value is converted to the struct
// field type when possible, otherwise an error is returned.
// The decoding of each struct field can be customized by the format string stored
// under the "chai" key stored in the struct field's tag.
// The content of the format string is used instead of the struct field name and passed
// to the GetByField method.
func StructScan(d types.Object, t interface{}) error {
	ref := reflect.ValueOf(t)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr {
		return errors.New("target must be pointer to a valid Go type")
	}

	if ref.Elem().Kind() != reflect.Struct {
		return errors.New("target must be pointer to a struct")
	}

	if ref.IsNil() {
		ref.Set(reflect.New(ref.Type().Elem()))
	}

	return structScan(d, ref)
}

func structScan(d types.Object, ref reflect.Value) error {
	if ref.Type().Implements(reflect.TypeOf((*Scanner)(nil)).Elem()) {
		return ref.Interface().(Scanner).ScanObject(d)
	}

	sref := reflect.Indirect(ref)
	stp := sref.Type()
	l := sref.NumField()
	for i := 0; i < l; i++ {
		f := sref.Field(i)
		sf := stp.Field(i)
		if sf.Anonymous {
			err := structScan(d, f)
			if err != nil {
				return err
			}
			continue
		}
		var name string
		if gtag, ok := sf.Tag.Lookup("chai"); ok {
			if gtag == "-" {
				continue
			}

			name = gtag
		} else {
			name = strings.ToLower(sf.Name)
		}
		v, err := d.GetByField(name)
		if errors.Is(err, types.ErrFieldNotFound) {
			v = types.NewNullValue()
		} else if err != nil {
			return err
		}

		if err := scanValue(v, f); err != nil {
			return err
		}
	}

	return nil
}

// SliceScan scans an array into a slice or fixed size array. t must be a pointer
// to a valid slice or array.
//
// It t is a slice pointer and its capacity is too low, a new slice will be allocated.
// Otherwise, its length is set to 0 so that its content is overwritten.
//
// If t is an array pointer, its capacity must be bigger than the length of a, otherwise an error is
// returned.
func SliceScan(a types.Array, t interface{}) error {
	return sliceScan(a, reflect.ValueOf(t))
}

func sliceScan(a types.Array, ref reflect.Value) error {
	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.IsNil() {
		return errors.New("target must be pointer to a slice or array")
	}

	tp := ref.Type()
	k := tp.Elem().Kind()
	if k != reflect.Array && k != reflect.Slice {
		return errors.New("target must be pointer to a slice or array")
	}

	al, err := ArrayLength(a)
	if err != nil {
		return err
	}

	sref := reflect.Indirect(ref)

	// if array, make sure it is big enough
	if k == reflect.Array && sref.Len() < al {
		return errors.New("array length too small")
	}

	// if slice, reduce its length to 0 to overwrite the buffer
	if k == reflect.Slice {
		if sref.Cap() < al {
			sref.Set(reflect.MakeSlice(tp.Elem(), 0, al))
		} else {
			sref.SetLen(0)
		}
	}

	stp := sref.Type()

	err = a.Iterate(func(i int, v types.Value) error {
		if k == reflect.Array {
			err := scanValue(v, sref.Index(i).Addr())
			if err != nil {
				return err
			}
		} else {
			newV := reflect.New(stp.Elem())

			err := scanValue(v, newV)
			if err != nil {
				return err
			}

			sref = reflect.Append(sref, reflect.Indirect(newV))
		}

		return nil
	})
	if err != nil {
		return err
	}

	if k == reflect.Slice {
		ref.Elem().Set(sref)
	}

	return nil
}

// MapScan decodes the object into a map.
func MapScan(d types.Object, t any) error {
	ref := reflect.ValueOf(t)
	if !ref.IsValid() {
		return &ErrUnsupportedType{ref, "t must be a valid reference"}
	}

	if ref.Kind() == reflect.Ptr {
		ref = reflect.Indirect(ref)
	}

	if ref.Kind() != reflect.Map {
		return &ErrUnsupportedType{ref, "t is not a map"}
	}

	return mapScan(d, ref)
}

func mapScan(d types.Object, ref reflect.Value) error {
	if ref.Type().Key().Kind() != reflect.String {
		return &ErrUnsupportedType{ref, "map key must be a string"}
	}

	if ref.IsNil() {
		ref.Set(reflect.MakeMap(ref.Type()))
	}

	return d.Iterate(func(f string, v types.Value) error {
		newV := reflect.New(ref.Type().Elem())

		err := scanValue(v, newV)
		if err != nil {
			return err
		}

		ref.SetMapIndex(reflect.ValueOf(f), newV.Elem())
		return nil
	})
}

// ScanValue scans v into t.
func ScanValue(v types.Value, t any) error {
	return scanValue(v, reflect.ValueOf(t))
}

func scanValue(v types.Value, ref reflect.Value) error {
	if !ref.IsValid() {
		return &ErrUnsupportedType{ref, "parameter is not a valid reference"}
	}

	if v.Type() == types.TypeNull {
		if ref.Type().Kind() != reflect.Ptr {
			return nil
		}

		if ref.IsNil() {
			return nil
		}

		if !ref.CanSet() {
			ref = reflect.Indirect(ref)
		}

		ref.Set(reflect.Zero(ref.Type()))
		return nil
	}

	if ref.Type().Kind() == reflect.Ptr && ref.IsNil() {
		ref.Set(reflect.New(ref.Type().Elem()))
	}

	ref = reflect.Indirect(ref)

	// if the user passed a **ptr
	// make sure it points to a valid value
	// or create one
	// then dereference
	if ref.Kind() == reflect.Ptr {
		if ref.IsNil() {
			ref.Set(reflect.New(ref.Type().Elem()))
		}

		ref = reflect.Indirect(ref)
	}

	// Scan nulls as Go zero values.
	if v.Type() == types.TypeNull {
		ref.Set(reflect.Zero(ref.Type()))
		return nil
	}

	switch ref.Kind() {
	case reflect.String:
		v, err := CastAsText(v)
		if err != nil {
			return err
		}
		ref.SetString(types.AsString(v))
		return nil
	case reflect.Bool:
		v, err := CastAsBool(v)
		if err != nil {
			return err
		}
		ref.SetBool(types.AsBool(v))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := CastAsInteger(v)
		if err != nil {
			return err
		}
		x := types.AsInt64(v)
		if x < 0 {
			return fmt.Errorf("cannot convert value %d into Go value of type %s", x, ref.Type().Name())
		}
		ref.SetUint(uint64(x))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := CastAsInteger(v)
		if err != nil {
			return err
		}
		ref.SetInt(types.AsInt64(v))
		return nil
	case reflect.Float32, reflect.Float64:
		v, err := CastAsDouble(v)
		if err != nil {
			return err
		}
		ref.SetFloat(types.AsFloat64(v))
		return nil
	case reflect.Interface:
		if !ref.IsNil() {
			return scanValue(v, ref.Elem())
		}
		switch v.Type() {
		case types.TypeObject:
			m := make(map[string]any)
			vm := reflect.ValueOf(m)
			ref.Set(vm)
			return mapScan(types.AsObject(v), vm)
		case types.TypeArray:
			var s []interface{}
			vs := reflect.ValueOf(&s)
			err := sliceScan(types.AsArray(v), vs)
			if err != nil {
				return err
			}
			ref.Set(vs.Elem())
			return nil
		case types.TypeText:
			// copy the string to avoid
			// keeping a reference to the underlying buffer
			// which could be reused
			cp := strings.Clone(types.AsString(v))
			ref.Set(reflect.ValueOf(cp))
			return nil
		case types.TypeBlob:
			// copy the byte slice to avoid
			// keeping a reference to the underlying buffer
			// which could be reused
			b := bytes.Clone(types.AsByteSlice(v))
			ref.Set(reflect.ValueOf(b))
			return nil
		}

		ref.Set(reflect.ValueOf(v.V()))
		return nil
	}

	// test with supported stdlib types
	switch ref.Type().String() {
	case "time.Time":
		switch v.Type() {
		case types.TypeText:
			parsed, err := time.Parse(time.RFC3339Nano, types.AsString(v))
			if err != nil {
				return err
			}

			ref.Set(reflect.ValueOf(parsed))
			return nil
		case types.TypeTimestamp:
			ref.Set(reflect.ValueOf(types.AsTime(v)))
			return nil
		}
	}

	switch ref.Kind() {
	case reflect.Struct:
		v, err := CastAsObject(v)
		if err != nil {
			return err
		}

		return structScan(types.AsObject(v), ref)
	case reflect.Slice:
		if ref.Type().Elem().Kind() == reflect.Uint8 {
			if v.Type() != types.TypeText && v.Type() != types.TypeBlob {
				return fmt.Errorf("cannot scan value of type %s to byte slice", v.Type())
			}
			if v.Type() == types.TypeText {
				ref.SetBytes([]byte(types.AsString(v)))
			} else {
				ref.SetBytes(types.AsByteSlice(v))
			}
			return nil
		}
		v, err := CastAsArray(v)
		if err != nil {
			return err
		}

		return sliceScan(types.AsArray(v), ref.Addr())
	case reflect.Array:
		if ref.Type().Elem().Kind() == reflect.Uint8 {
			if v.Type() != types.TypeText && v.Type() != types.TypeBlob {
				return fmt.Errorf("cannot scan value of type %s to byte slice", v.Type())
			}
			reflect.Copy(ref, reflect.ValueOf(v.V()))
			return nil
		}
		v, err := CastAsArray(v)
		if err != nil {
			return err
		}

		return sliceScan(types.AsArray(v), ref.Addr())
	case reflect.Map:
		v, err := CastAsObject(v)
		if err != nil {
			return err
		}

		return mapScan(types.AsObject(v), ref)
	}

	return &ErrUnsupportedType{ref, "Invalid type"}
}

// ScanRow scans a row into dest which must be either a struct pointer, a map or a map pointer.
func ScanRow(d types.Object, t interface{}) error {
	ref := reflect.ValueOf(t)

	if !ref.IsValid() {
		return errors.New("target must be pointer to a valid Go type")
	}

	switch reflect.Indirect(ref).Kind() {
	case reflect.Map:
		return mapScan(d, ref)
	case reflect.Struct:
		if ref.IsNil() {
			ref.Set(reflect.New(ref.Type().Elem()))
		}
		return structScan(d, ref)
	default:
		return errors.New("target must be a either a pointer to struct, a map or a map pointer")
	}
}

// ScanIterator scans a row iterator into a slice or fixed size array. t must be a pointer
// to a valid slice or array.
//
// It t is a slice pointer and its capacity is too low, a new slice will be allocated.
// Otherwise, its length is set to 0 so that its content is overwritten.
//
// If t is an array pointer, its capacity must be bigger than the length of a, otherwise an error is
// returned.
func ScanIterator(it Iterator, t interface{}) error {
	a := iteratorArray{it: it}
	return SliceScan(&a, t)
}

type iteratorArray struct {
	it Iterator
}

func (it *iteratorArray) Iterate(fn func(i int, value types.Value) error) error {
	count := 0
	return it.it.Iterate(func(d types.Object) error {
		err := fn(count, types.NewObjectValue(d))
		if err != nil {
			return err
		}
		count++
		return nil
	})
}

func (it *iteratorArray) GetByIndex(i int) (types.Value, error) {
	panic("not implemented")
}

func (it *iteratorArray) MarshalJSON() ([]byte, error) {
	return MarshalJSONArray(it)
}

// ScanField scans a single field into dest.
func ScanField(d types.Object, field string, dest interface{}) error {
	v, err := d.GetByField(field)
	if err != nil {
		return err
	}

	return ScanValue(v, dest)
}

// ScanPath scans a single path into dest.
func ScanPath(d types.Object, path Path, dest interface{}) error {
	v, err := path.GetValueFromObject(d)
	if err != nil {
		return err
	}

	return ScanValue(v, dest)
}
