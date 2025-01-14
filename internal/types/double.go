package types

import (
	"math"
	"strconv"
)

var _ Numeric = NewDoubleValue(0)

type DoubleValue float64

// NewDoubleValue returns a SQL DOUBLE value.
func NewDoubleValue(x float64) DoubleValue {
	return DoubleValue(x)
}

func (v DoubleValue) V() any {
	return float64(v)
}

func (v DoubleValue) Type() Type {
	return TypeDouble
}

func (v DoubleValue) IsZero() (bool, error) {
	return v == 0, nil
}

func (v DoubleValue) String() string {
	f := AsFloat64(v)
	abs := math.Abs(f)
	fmt := byte('f')
	if abs != 0 {
		if abs < 1e-6 || abs >= 1e15 {
			fmt = 'e'
		}
	}

	// By default the precision is -1 to use the smallest number of digits.
	// See https://pkg.go.dev/strconv#FormatFloat
	prec := -1
	// if the number is round, add .0
	if float64(int64(f)) == f {
		prec = 1
	}
	return strconv.FormatFloat(AsFloat64(v), fmt, prec, 64)
}

func (v DoubleValue) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v DoubleValue) MarshalJSON() ([]byte, error) {
	f := AsFloat64(v)
	abs := math.Abs(f)
	fmt := byte('f')
	if abs != 0 {
		if abs < 1e-6 || abs >= 1e15 {
			fmt = 'e'
		}
	}

	// By default the precision is -1 to use the smallest number of digits.
	// See https://pkg.go.dev/strconv#FormatFloat
	prec := -1
	return strconv.AppendFloat(nil, AsFloat64(v), fmt, prec, 64), nil
}

func (v DoubleValue) EQ(other Value) (bool, error) {
	t := other.Type()
	switch t {
	case TypeDouble:
		return float64(v) == AsFloat64(other), nil
	case TypeInteger:
		return float64(v) == float64(AsInt64(other)), nil
	default:
		return false, nil
	}
}

func (v DoubleValue) GT(other Value) (bool, error) {
	t := other.Type()
	switch t {
	case TypeDouble:
		return float64(v) > AsFloat64(other), nil
	case TypeInteger:
		return float64(v) > float64(AsInt64(other)), nil
	default:
		return false, nil
	}
}

func (v DoubleValue) GTE(other Value) (bool, error) {
	t := other.Type()
	switch t {
	case TypeDouble:
		return float64(v) >= AsFloat64(other), nil
	case TypeInteger:
		return float64(v) >= float64(AsInt64(other)), nil
	default:
		return false, nil
	}
}

func (v DoubleValue) LT(other Value) (bool, error) {
	t := other.Type()
	switch t {
	case TypeDouble:
		return float64(v) < AsFloat64(other), nil
	case TypeInteger:
		return float64(v) < float64(AsInt64(other)), nil
	default:
		return false, nil
	}
}

func (v DoubleValue) LTE(other Value) (bool, error) {
	t := other.Type()
	switch t {
	case TypeDouble:
		return float64(v) <= AsFloat64(other), nil
	case TypeInteger:
		return float64(v) <= float64(AsInt64(other)), nil
	default:
		return false, nil
	}
}

func (v DoubleValue) Between(a, b Value) (bool, error) {
	if !a.Type().IsNumber() || !b.Type().IsNumber() {
		return false, nil
	}

	ok, err := a.LTE(v)
	if err != nil || !ok {
		return false, err
	}

	return b.GTE(v)
}

func (v DoubleValue) Add(other Numeric) (Value, error) {
	switch other.Type() {
	case TypeInteger:
		return NewDoubleValue(float64(v) + float64(AsInt64(other))), nil
	case TypeDouble:
		return NewDoubleValue(float64(v) + AsFloat64(other)), nil
	}

	return NewNullValue(), nil
}

func (v DoubleValue) Sub(other Numeric) (Value, error) {
	switch other.Type() {
	case TypeInteger:
		return NewDoubleValue(float64(v) - float64(AsInt64(other))), nil
	case TypeDouble:
		return NewDoubleValue(float64(v) - AsFloat64(other)), nil
	}

	return NewNullValue(), nil
}

func (v DoubleValue) Mul(other Numeric) (Value, error) {
	switch other.Type() {
	case TypeInteger:
		return NewDoubleValue(float64(v) * float64(AsInt64(other))), nil
	case TypeDouble:
		return NewDoubleValue(float64(v) * AsFloat64(other)), nil
	}

	return NewNullValue(), nil
}

func (v DoubleValue) Div(other Numeric) (Value, error) {
	switch other.Type() {
	case TypeInteger:
		xb := float64(AsInt64(other))
		if xb == 0 {
			return NewNullValue(), nil
		}

		return NewDoubleValue(float64(v) / xb), nil
	case TypeDouble:
		xb := AsFloat64(other)
		if xb == 0 {
			return NewNullValue(), nil
		}

		return NewDoubleValue(float64(v) / xb), nil
	}

	return NewNullValue(), nil
}

func (v DoubleValue) Mod(other Numeric) (Value, error) {
	switch other.Type() {
	case TypeInteger:
		xb := float64(AsInt64(other))
		xr := math.Mod(float64(v), xb)
		if math.IsNaN(xr) {
			return NewNullValue(), nil
		}

		return NewDoubleValue(xr), nil
	case TypeDouble:
		xb := AsFloat64(other)
		xr := math.Mod(float64(v), xb)
		if math.IsNaN(xr) {
			return NewNullValue(), nil
		}

		return NewDoubleValue(xr), nil
	}

	return NewNullValue(), nil
}

func (v DoubleValue) BitwiseAnd(other Numeric) (Value, error) {
	switch other.Type() {
	case TypeInteger:
		return NewIntegerValue(int64(v) & AsInt64(other)), nil
	case TypeDouble:
		xa := int64(v)
		xb := int64(AsFloat64(other))
		return NewIntegerValue(xa & xb), nil
	}

	return NewNullValue(), nil
}

func (v DoubleValue) BitwiseOr(other Numeric) (Value, error) {
	switch other.Type() {
	case TypeInteger:
		return NewIntegerValue(int64(v) | AsInt64(other)), nil
	case TypeDouble:
		xa := int64(v)
		xb := int64(AsFloat64(other))
		return NewIntegerValue(xa | xb), nil
	}

	return NewNullValue(), nil
}

func (v DoubleValue) BitwiseXor(other Numeric) (Value, error) {
	switch other.Type() {
	case TypeInteger:
		return NewIntegerValue(int64(v) ^ AsInt64(other)), nil
	case TypeDouble:
		xa := int64(v)
		xb := int64(AsFloat64(other))
		return NewIntegerValue(xa ^ xb), nil
	}

	return NewNullValue(), nil
}
