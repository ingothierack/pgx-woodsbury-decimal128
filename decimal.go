package decimal

import (
	"fmt"
	"math"
	"math/big"
	"reflect"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/woodsbury/decimal128"
)

type Decimal decimal128.Decimal

func (d *Decimal) ScanNumeric(v pgtype.Numeric) error {
	if !v.Valid {
		return fmt.Errorf("cannot scan Null into *decimal128.Decimal")
	}

	if v.NaN {
		return fmt.Errorf("cannot scan NaN into *decimal128.Decimal")
	}

	if v.InfinityModifier != pgtype.Finite {
		return fmt.Errorf("cannot scan %v into *decimal128.Decimal", v.InfinityModifier)
	}

	if v.Int.BitLen() > 128 {
		panic("bitlen outside of range")
	}

	dec := decimal128.Decimal{}

	if v.Int.Sign() < 0 {
		dec.Compose(0, true, v.Int.Bytes(), v.Exp)
	} else {
		dec.Compose(0, false, v.Int.Bytes(), v.Exp)
	}

	*d = Decimal(dec)

	return nil
}

func (d Decimal) NumericValue() (pgtype.Numeric, error) {
	dd := decimal128.Decimal(d)
	_, sign, value, exp := dd.Decompose(nil)

	z := new(big.Int)
	if sign {
		z.SetBytes(value)
		z = z.Neg(z)
	} else {
		z.SetBytes(value)
	}
	return pgtype.Numeric{Int: z, Exp: exp, Valid: true}, nil
}

func (d *Decimal) ScanFloat64(v pgtype.Float8) error {
	if !v.Valid {
		return fmt.Errorf("cannot scan NULL into *decimal.Decimal")
	}

	if math.IsNaN(v.Float64) {
		return fmt.Errorf("cannot scan NaN into *decimal.Decimal")
	}

	if math.IsInf(v.Float64, 0) {
		return fmt.Errorf("cannot scan %v into *decimal.Decimal", v.Float64)
	}
	*d = Decimal(decimal128.FromFloat64(v.Float64))
	return nil
}

func (d Decimal) Int64Value() (pgtype.Int8, error) {
	dd := decimal128.Decimal(d)
	// buf := make([]byte, 1024)

	// form, sign, value, exp := dd.Decompose(buf)

	valint64, valid := dd.Int64()

	if !valid {
		return pgtype.Int8{}, fmt.Errorf("cannot convert %v to int64", dd)
	}

	return pgtype.Int8{Int64: valint64, Valid: true}, nil
}

type NullDecimal struct {
	Decimal decimal128.Decimal
	Valid   bool
}

func (d *NullDecimal) ScanNumeric(v pgtype.Numeric) error {
	if !v.Valid {
		*d = NullDecimal{}
		return nil
	}

	if v.NaN {
		return fmt.Errorf("cannot scan NaN into *decimal.NullDecimal")
	}

	if v.InfinityModifier != pgtype.Finite {
		return fmt.Errorf("cannot scan %v into *decimal.NullDecimal", v.InfinityModifier)
	}

	dec := decimal128.Decimal{}

	if v.Int.Sign() < 0 {
		dec.Compose(0, true, v.Int.Bytes(), v.Exp)
	} else {
		dec.Compose(0, false, v.Int.Bytes(), v.Exp)
	}
	*d = NullDecimal(NullDecimal{Decimal: dec, Valid: true})
	// *d = NullDecimal(dec)
	return nil
}

func (d NullDecimal) NumericValue() (pgtype.Numeric, error) {
	if !d.Valid {
		return pgtype.Numeric{}, nil
	}

	dd := decimal128.Decimal(d.Decimal)
	if dd.IsNaN() {
		return pgtype.Numeric{}, nil
	}
	buf := make([]byte, 1024)

	_, sign, value, exp := dd.Decompose(buf)

	z := new(big.Int)
	if sign {
		z.SetBytes(value)
		z = z.Neg(z)
	} else {
		z.SetBytes(value)
	}
	return pgtype.Numeric{Int: z, Exp: exp, Valid: true}, nil
}

func (d *NullDecimal) ScanFloat64(v pgtype.Float8) error {
	if !v.Valid {
		*d = NullDecimal{}
		return nil
	}

	if math.IsNaN(v.Float64) {
		return fmt.Errorf("cannot scan NaN into *decimal.NullDecimal")
	}

	if math.IsInf(v.Float64, 0) {
		return fmt.Errorf("cannot scan %v into *decimal.NullDecimal", v.Float64)
	}

	*d = NullDecimal(NullDecimal{Decimal: decimal128.FromFloat64(v.Float64), Valid: true})

	return nil
}

func (d NullDecimal) Float64Value() (pgtype.Float8, error) {
	if !d.Valid {
		return pgtype.Float8{}, nil
	}

	dd := NullDecimal(d)
	return pgtype.Float8{Float64: dd.Decimal.Float64(), Valid: true}, nil
}

func (d *NullDecimal) ScanInt64(v pgtype.Int8) error {
	if !v.Valid {
		*d = NullDecimal{}
		return nil
	}

	*d = NullDecimal(NullDecimal{Decimal: decimal128.FromInt64(v.Int64), Valid: true})

	return nil
}

func (d NullDecimal) Int64Value() (pgtype.Int8, error) {
	if !d.Valid {
		return pgtype.Int8{}, nil
	}

	if d.Decimal.IsNaN() {
		return pgtype.Int8{}, nil
	}

	bi := NullDecimal(d).Decimal.Int(nil)
	if !bi.IsInt64() {
		return pgtype.Int8{}, fmt.Errorf("cannot convert %v to int64", d)
	}

	return pgtype.Int8{Int64: bi.Int64(), Valid: true}, nil
}

func TryWrapNumericEncodePlan(value interface{}) (plan pgtype.WrappedEncodePlanNextSetter, nextValue interface{}, ok bool) {
	switch value := value.(type) {
	case decimal128.Decimal:
		return &wrapDecimalEncodePlan{}, Decimal(value), true
		// case decimal128.NullDecimal:
		// 	return &wrapNullDecimalEncodePlan{}, NullDecimal(value), true
	}

	return nil, nil, false
}

type wrapDecimalEncodePlan struct {
	next pgtype.EncodePlan
}

func (plan *wrapDecimalEncodePlan) SetNext(next pgtype.EncodePlan) { plan.next = next }

func (plan *wrapDecimalEncodePlan) Encode(value interface{}, buf []byte) (newBuf []byte, err error) {
	return plan.next.Encode(Decimal(value.(decimal128.Decimal)), buf)
}

type wrapNullDecimalEncodePlan struct {
	next pgtype.EncodePlan
}

func (plan *wrapNullDecimalEncodePlan) SetNext(next pgtype.EncodePlan) { plan.next = next }

func (plan *wrapNullDecimalEncodePlan) Encode(value interface{}, buf []byte) (newBuf []byte, err error) {
	return plan.next.Encode(NullDecimal(value.(NullDecimal)), buf)
}

func TryWrapNumericScanPlan(target interface{}) (plan pgtype.WrappedScanPlanNextSetter, nextDst interface{}, ok bool) {
	switch target := target.(type) {
	case *decimal128.Decimal:
		return &wrapDecimalScanPlan{}, (*Decimal)(target), true
		// case *decimal128.Decimal:
		// 	return &wrapNullDecimalScanPlan{}, (*NullDecimal)(target), true
	}

	return nil, nil, false
}

type wrapDecimalScanPlan struct {
	next pgtype.ScanPlan
}

func (plan *wrapDecimalScanPlan) SetNext(next pgtype.ScanPlan) { plan.next = next }

func (plan *wrapDecimalScanPlan) Scan(src []byte, dst interface{}) error {
	return plan.next.Scan(src, (*Decimal)(dst.(*decimal128.Decimal)))
}

type wrapNullDecimalScanPlan struct {
	next pgtype.ScanPlan
}

func (plan *wrapNullDecimalScanPlan) SetNext(next pgtype.ScanPlan) { plan.next = next }

func (plan *wrapNullDecimalScanPlan) Scan(src []byte, dst interface{}) error {
	return plan.next.Scan(src, (*NullDecimal)(dst.(*NullDecimal)))
}

type NumericCodec struct {
	pgtype.NumericCodec
}

func (NumericCodec) DecodeValue(tm *pgtype.Map, oid uint32, format int16, src []byte) (interface{}, error) {
	if src == nil {
		return nil, nil
	}

	var target decimal128.Decimal
	scanPlan := tm.PlanScan(oid, format, &target)
	if scanPlan == nil {
		return nil, fmt.Errorf("PlanScan did not find a plan")
	}

	err := scanPlan.Scan(src, &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

// Register registers the shopspring/decimal integration with a pgtype.ConnInfo.
func Register(m *pgtype.Map) {
	m.TryWrapEncodePlanFuncs = append([]pgtype.TryWrapEncodePlanFunc{TryWrapNumericEncodePlan}, m.TryWrapEncodePlanFuncs...)
	m.TryWrapScanPlanFuncs = append([]pgtype.TryWrapScanPlanFunc{TryWrapNumericScanPlan}, m.TryWrapScanPlanFuncs...)

	m.RegisterType(&pgtype.Type{
		Name:  "numeric",
		OID:   pgtype.NumericOID,
		Codec: NumericCodec{},
	})

	registerDefaultPgTypeVariants := func(name, arrayName string, value interface{}) {
		// T
		m.RegisterDefaultPgType(value, name)

		// *T
		valueType := reflect.TypeOf(value)
		m.RegisterDefaultPgType(reflect.New(valueType).Interface(), name)

		// []T
		sliceType := reflect.SliceOf(valueType)
		m.RegisterDefaultPgType(reflect.MakeSlice(sliceType, 0, 0).Interface(), arrayName)

		// *[]T
		m.RegisterDefaultPgType(reflect.New(sliceType).Interface(), arrayName)

		// []*T
		sliceOfPointerType := reflect.SliceOf(reflect.TypeOf(reflect.New(valueType).Interface()))
		m.RegisterDefaultPgType(reflect.MakeSlice(sliceOfPointerType, 0, 0).Interface(), arrayName)

		// *[]*T
		m.RegisterDefaultPgType(reflect.New(sliceOfPointerType).Interface(), arrayName)
	}

	registerDefaultPgTypeVariants("numeric", "_numeric", decimal128.Decimal{})
	registerDefaultPgTypeVariants("numeric", "_numeric", decimal128.Decimal{})
	registerDefaultPgTypeVariants("numeric", "_numeric", Decimal{})
	registerDefaultPgTypeVariants("numeric", "_numeric", NullDecimal{})
}
