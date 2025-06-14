package decimal

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"sync"

	"github.com/ingothierack/decimal128"
	"github.com/jackc/pgx/v5/pgtype"
)

type Decimal decimal128.Decimal

const ErrScanNull = "cannot scan NULL into *decimal128.Decimal"
const ErrScanNaN = "cannot scan NaN into *decimal128.Decimal"
const ErrScanInf = "cannot scan %v into *decimal128.Decimal"

// Predeclare common errors
var (
	errScanNull = fmt.Errorf(ErrScanNull)
	errScanNaN  = fmt.Errorf(ErrScanNaN)
)

// Use a sync.Pool for buffers
var bufPool = sync.Pool{
	New: func() any {
		return make([]byte, 1024)
	},
}

func (d *Decimal) ScanNumeric(v pgtype.Numeric) error {
	if !v.Valid {
		return errScanNull
	}

	if v.NaN {
		return errScanNaN
	}

	if v.InfinityModifier != pgtype.Finite {
		return fmt.Errorf(ErrScanInf, v.InfinityModifier)
	}

	if v.Int.BitLen() > 128 {
		panic("bitlen outside of range")
	}

	*d = Decimal(composeDecimal(v))

	return nil
}

// description: This function returns the value of the Decimal type
// return: pgtype.Numeric, error
func (d Decimal) NumericValue() (pgtype.Numeric, error) {
	var inf pgtype.InfinityModifier

	dd := decimal128.Decimal(d)
	modifier, sign, value, exp := dd.Decompose(nil)

	switch modifier {
	case 0:
		inf = pgtype.Finite
	case 1:
		inf = pgtype.Infinity
	case 2:
		if sign {
			inf = pgtype.NegativeInfinity
		}
	default:
		inf = pgtype.Finite
	}

	z := new(big.Int).SetBytes(value)
	if sign {
		z = z.Neg(z)
	}

	return pgtype.Numeric{Int: z, Exp: exp, Valid: true, NaN: false, InfinityModifier: inf}, nil
}

func (d *Decimal) ScanFloat64(v pgtype.Float8) error {
	if !v.Valid {
		return errScanNull
	}

	floatVal := v.Float64

	if math.IsNaN(floatVal) {
		return errScanNaN
	}

	if math.IsInf(floatVal, 0) {
		return fmt.Errorf(ErrScanInf, floatVal)
	}

	s := strconv.FormatFloat(floatVal, 'f', -1, 64)
	*d = Decimal(decimal128.MustParse(s))
	return nil
}

func (d Decimal) Float64Value() (pgtype.Float8, error) {
	dd := decimal128.Decimal(d)
	return pgtype.Float8{Float64: dd.Float64(), Valid: true}, nil
}

func (d *Decimal) ScanInt64(v pgtype.Int8) error {
	if !v.Valid {
		*d = Decimal{}
		return nil
	}

	*d = Decimal(decimal128.FromInt64(v.Int64))

	return nil
}

func (d Decimal) Int64Value() (pgtype.Int8, error) {
	dd := decimal128.Decimal(d)
	if dd.IsNaN() {
		return pgtype.Int8{Int64: 0, Valid: false}, nil
	}

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
		return fmt.Errorf(ErrScanNaN)
	}

	if v.InfinityModifier != pgtype.Finite {
		return fmt.Errorf(ErrScanInf, v.InfinityModifier)
	}

	*d = NullDecimal{Decimal: composeDecimal(v), Valid: true}
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
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	_, sign, value, exp := dd.Decompose(buf)

	z := new(big.Int).SetBytes(value)
	if sign {
		z = z.Neg(z)
	}

	return pgtype.Numeric{Int: z, Exp: exp, Valid: true}, nil
}

func (d *NullDecimal) ScanFloat64(v pgtype.Float8) error {
	if !v.Valid {
		*d = NullDecimal{}
		return nil
	}

	if math.IsNaN(v.Float64) {
		return errScanNaN
	}

	if math.IsInf(v.Float64, 0) {
		return fmt.Errorf(ErrScanInf, v.Float64)
	}

	s := strconv.FormatFloat(v.Float64, 'f', -1, 64)
	*d = NullDecimal{Decimal: decimal128.MustParse(s), Valid: true}
	return nil

}

func (d NullDecimal) Float64Value() (pgtype.Float8, error) {
	if !d.Valid {
		return pgtype.Float8{}, nil
	}
	dd := d.Decimal

	return pgtype.Float8{Float64: dd.Float64(), Valid: true}, nil
}

func (d *NullDecimal) ScanInt64(v pgtype.Int8) error {
	if !v.Valid {
		*d = NullDecimal{}
		return nil
	}

	*d = NullDecimal{Decimal: decimal128.FromInt64(v.Int64), Valid: true}

	return nil
}

func (d NullDecimal) Int64Value() (pgtype.Int8, error) {
	if !d.Valid {
		return pgtype.Int8{}, nil
	}

	if d.Decimal.IsNaN() {
		return pgtype.Int8{}, nil
	}

	bi := d.Decimal.Int(nil)
	if !bi.IsInt64() {
		return pgtype.Int8{}, fmt.Errorf("cannot convert %v to int64", d)
	}

	return pgtype.Int8{Int64: bi.Int64(), Valid: true}, nil
}

func TryWrapNumericEncodePlan(value any) (plan pgtype.WrappedEncodePlanNextSetter, nextValue any, ok bool) {
	switch value := value.(type) {
	case decimal128.Decimal:
		return &wrapDecimalEncodePlan{}, Decimal(value), true
	case NullDecimal:
		return &wrapNullDecimalEncodePlan{}, NullDecimal(value), true
	}

	return nil, nil, false
}

type wrapDecimalEncodePlan struct {
	next pgtype.EncodePlan
}

func (plan *wrapDecimalEncodePlan) SetNext(next pgtype.EncodePlan) {
	plan.next = next
}

func (plan *wrapDecimalEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	return plan.next.Encode(Decimal(value.(decimal128.Decimal)), buf)
}

type wrapNullDecimalEncodePlan struct {
	next pgtype.EncodePlan
}

func (plan *wrapNullDecimalEncodePlan) SetNext(next pgtype.EncodePlan) {
	plan.next = next
}

func (plan *wrapNullDecimalEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	return plan.next.Encode(NullDecimal(value.(NullDecimal)), buf)
}

func TryWrapNumericScanPlan(target any) (plan pgtype.WrappedScanPlanNextSetter, nextDst any, ok bool) {
	switch target := target.(type) {
	case *decimal128.Decimal:
		return &wrapDecimalScanPlan{}, (*Decimal)(target), true
	case *NullDecimal:
		return &wrapNullDecimalScanPlan{}, (*NullDecimal)(target), true
	}

	return nil, nil, false
}

type wrapDecimalScanPlan struct {
	next pgtype.ScanPlan
}

func (plan *wrapDecimalScanPlan) SetNext(next pgtype.ScanPlan) { plan.next = next }

func (plan *wrapDecimalScanPlan) Scan(src []byte, dst any) error {
	return plan.next.Scan(src, (*Decimal)(dst.(*decimal128.Decimal)))
}

type wrapNullDecimalScanPlan struct {
	next pgtype.ScanPlan
}

func (plan *wrapNullDecimalScanPlan) SetNext(next pgtype.ScanPlan) {
	plan.next = next
}

func (plan *wrapNullDecimalScanPlan) Scan(src []byte, dst any) error {
	return plan.next.Scan(src, (*NullDecimal)(dst.(*NullDecimal)))
}

type NumericCodec struct {
	pgtype.NumericCodec
}

func (NumericCodec) DecodeValue(tm *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
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

// Register registers the decimal integration with a pgtype.ConnInfo.
func Register(m *pgtype.Map) {
	m.TryWrapEncodePlanFuncs = append([]pgtype.TryWrapEncodePlanFunc{TryWrapNumericEncodePlan}, m.TryWrapEncodePlanFuncs...)
	m.TryWrapScanPlanFuncs = append([]pgtype.TryWrapScanPlanFunc{TryWrapNumericScanPlan}, m.TryWrapScanPlanFuncs...)

	m.RegisterType(&pgtype.Type{
		Name:  "numeric",
		OID:   pgtype.NumericOID,
		Codec: NumericCodec{},
	})

	registerDefaultPgTypeVariants := func(name, arrayName string, value any) {
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
	registerDefaultPgTypeVariants("numeric", "_numeric", Decimal{})
	registerDefaultPgTypeVariants("numeric", "_numeric", NullDecimal{})
}

func composeDecimal(v pgtype.Numeric) decimal128.Decimal {
	dec := decimal128.Decimal{}
	if v.Int.Sign() < 0 {
		dec.Compose(0, true, v.Int.Bytes(), v.Exp)
	} else {
		dec.Compose(0, false, v.Int.Bytes(), v.Exp)
	}
	return dec
}
