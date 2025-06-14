package decimal_test

import (
	"context"
	"math"
	"math/big"
	"testing"

	"github.com/ingothierack/decimal128"
	pgxdecimal "github.com/ingothierack/pgx-woodsbury-decimal128"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Benchmark data
var (
	benchmarkDecimal128Values = []decimal128.Decimal{
		decimal128.MustParse("0"),
		decimal128.MustParse("1"),
		decimal128.MustParse("-1"),
		decimal128.MustParse("123.456"),
		decimal128.MustParse("-123.456"),
		decimal128.MustParse("1234567890.123456789"),
		decimal128.MustParse("-1234567890.123456789"),
		decimal128.MustParse("0.000000001"),
		decimal128.MustParse("999999999999999999.999999999999999"),
		decimal128.FromInt64(math.MaxInt64),
		decimal128.FromInt64(math.MinInt64),
	}

	benchmarkNumericValues []pgtype.Numeric

	benchmarkFloat64Values = []pgtype.Float8{
		{Float64: 0.0, Valid: true},
		{Float64: 1.0, Valid: true},
		{Float64: -1.0, Valid: true},
		{Float64: 123.456, Valid: true},
		{Float64: -123.456, Valid: true},
		{Float64: 1234567890.123456789, Valid: true},
		{Float64: -1234567890.123456789, Valid: true},
		{Float64: 0.000000001, Valid: true},
		{Float64: math.MaxFloat64, Valid: true},
		{Valid: false}, // NULL value
	}

	benchmarkInt64Values = []pgtype.Int8{
		{Int64: 0, Valid: true},
		{Int64: 1, Valid: true},
		{Int64: -1, Valid: true},
		{Int64: 123456, Valid: true},
		{Int64: -123456, Valid: true},
		{Int64: math.MaxInt64, Valid: true},
		{Int64: math.MinInt64, Valid: true},
		{Valid: false}, // NULL value
	}
)

func init() {
	// Initialize pgtype.Numeric values properly
	benchmarkNumericValues = make([]pgtype.Numeric, len(benchmarkDecimal128Values)+1)
	for i, val := range benchmarkDecimal128Values {
		dd := val
		_, sign, value, exp := dd.Decompose(nil)
		z := new(big.Int).SetBytes(value)
		if sign {
			z = z.Neg(z)
		}
		benchmarkNumericValues[i] = pgtype.Numeric{
			Int:   z,
			Exp:   exp,
			Valid: true,
		}
	}
	// Add NULL value
	benchmarkNumericValues[len(benchmarkNumericValues)-1] = pgtype.Numeric{Valid: false}
}

// Benchmark Decimal.ScanNumeric
func BenchmarkDecimalScanNumeric(b *testing.B) {
	var d pgxdecimal.Decimal

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		numeric := benchmarkNumericValues[i%len(benchmarkNumericValues)]
		d.ScanNumeric(numeric)
	}
}

// Benchmark Decimal.NumericValue
func BenchmarkDecimalNumericValue(b *testing.B) {
	decimals := make([]pgxdecimal.Decimal, len(benchmarkDecimal128Values))
	for i, val := range benchmarkDecimal128Values {
		decimals[i] = pgxdecimal.Decimal(val)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := decimals[i%len(decimals)]
		d.NumericValue()
	}
}

// Benchmark Decimal.ScanFloat64
func BenchmarkDecimalScanFloat64(b *testing.B) {
	var d pgxdecimal.Decimal

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		float8 := benchmarkFloat64Values[i%len(benchmarkFloat64Values)]
		d.ScanFloat64(float8)
	}
}

// Benchmark Decimal.Float64Value
func BenchmarkDecimalFloat64Value(b *testing.B) {
	decimals := make([]pgxdecimal.Decimal, len(benchmarkDecimal128Values))
	for i, val := range benchmarkDecimal128Values {
		decimals[i] = pgxdecimal.Decimal(val)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := decimals[i%len(decimals)]
		d.Float64Value()
	}
}

// Benchmark Decimal.ScanInt64
func BenchmarkDecimalScanInt64(b *testing.B) {
	var d pgxdecimal.Decimal

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		int8 := benchmarkInt64Values[i%len(benchmarkInt64Values)]
		d.ScanInt64(int8)
	}
}

// Benchmark Decimal.Int64Value
func BenchmarkDecimalInt64Value(b *testing.B) {
	decimals := make([]pgxdecimal.Decimal, len(benchmarkDecimal128Values))
	for i, val := range benchmarkDecimal128Values {
		decimals[i] = pgxdecimal.Decimal(val)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := decimals[i%len(decimals)]
		d.Int64Value()
	}
}

// Benchmark NullDecimal.ScanNumeric
func BenchmarkNullDecimalScanNumeric(b *testing.B) {
	var d pgxdecimal.NullDecimal

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		numeric := benchmarkNumericValues[i%len(benchmarkNumericValues)]
		d.ScanNumeric(numeric)
	}
}

// Benchmark NullDecimal.NumericValue
func BenchmarkNullDecimalNumericValue(b *testing.B) {
	nullDecimals := make([]pgxdecimal.NullDecimal, len(benchmarkDecimal128Values)+1)
	for i, val := range benchmarkDecimal128Values {
		nullDecimals[i] = pgxdecimal.NullDecimal{Decimal: val, Valid: true}
	}
	nullDecimals[len(nullDecimals)-1] = pgxdecimal.NullDecimal{Valid: false} // NULL value

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := nullDecimals[i%len(nullDecimals)]
		d.NumericValue()
	}
}

// Benchmark NullDecimal.ScanFloat64
func BenchmarkNullDecimalScanFloat64(b *testing.B) {
	var d pgxdecimal.NullDecimal

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		float8 := benchmarkFloat64Values[i%len(benchmarkFloat64Values)]
		d.ScanFloat64(float8)
	}
}

// Benchmark NullDecimal.Float64Value
func BenchmarkNullDecimalFloat64Value(b *testing.B) {
	nullDecimals := make([]pgxdecimal.NullDecimal, len(benchmarkDecimal128Values)+1)
	for i, val := range benchmarkDecimal128Values {
		nullDecimals[i] = pgxdecimal.NullDecimal{Decimal: val, Valid: true}
	}
	nullDecimals[len(nullDecimals)-1] = pgxdecimal.NullDecimal{Valid: false} // NULL value

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := nullDecimals[i%len(nullDecimals)]
		d.Float64Value()
	}
}

// Benchmark NullDecimal.ScanInt64
func BenchmarkNullDecimalScanInt64(b *testing.B) {
	var d pgxdecimal.NullDecimal

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		int8 := benchmarkInt64Values[i%len(benchmarkInt64Values)]
		d.ScanInt64(int8)
	}
}

// Benchmark NullDecimal.Int64Value
func BenchmarkNullDecimalInt64Value(b *testing.B) {
	nullDecimals := make([]pgxdecimal.NullDecimal, len(benchmarkDecimal128Values)+1)
	for i, val := range benchmarkDecimal128Values {
		nullDecimals[i] = pgxdecimal.NullDecimal{Decimal: val, Valid: true}
	}
	nullDecimals[len(nullDecimals)-1] = pgxdecimal.NullDecimal{Valid: false} // NULL value

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := nullDecimals[i%len(nullDecimals)]
		d.Int64Value()
	}
}

// Benchmark TryWrapNumericEncodePlan
func BenchmarkTryWrapNumericEncodePlan(b *testing.B) {
	values := make([]interface{}, len(benchmarkDecimal128Values)*2)
	for i, val := range benchmarkDecimal128Values {
		values[i*2] = val
		values[i*2+1] = pgxdecimal.NullDecimal{Decimal: val, Valid: true}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val := values[i%len(values)]
		pgxdecimal.TryWrapNumericEncodePlan(val)
	}
}

// Benchmark TryWrapNumericScanPlan
func BenchmarkTryWrapNumericScanPlan(b *testing.B) {
	targets := make([]interface{}, len(benchmarkDecimal128Values)*2)
	for i := range benchmarkDecimal128Values {
		var d decimal128.Decimal
		var nd pgxdecimal.NullDecimal
		targets[i*2] = &d
		targets[i*2+1] = &nd
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target := targets[i%len(targets)]
		pgxdecimal.TryWrapNumericScanPlan(target)
	}
}

// Benchmark decimal composition through ScanNumeric
func BenchmarkDecimalComposition(b *testing.B) {
	var d pgxdecimal.Decimal

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		numeric := benchmarkNumericValues[i%len(benchmarkNumericValues)]
		if numeric.Valid {
			d.ScanNumeric(numeric)
		}
	}
}

// Benchmark with database operations (requires database connection)
func BenchmarkDatabaseRoundTrip(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping database benchmark in short mode")
	}

	ctx := context.Background()
	defaultConnTestRunner.RunTest(ctx, b, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		original := decimal128.MustParse("123.456789")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result decimal128.Decimal
			err := conn.QueryRow(ctx, "SELECT $1::numeric", original).Scan(&result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Benchmark array operations
func BenchmarkArrayOperations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping database benchmark in short mode")
	}

	ctx := context.Background()
	defaultConnTestRunner.RunTest(ctx, b, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		inputSlice := make([]decimal128.Decimal, 100)
		for i := range inputSlice {
			inputSlice[i] = decimal128.FromInt64(int64(i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var outputSlice []decimal128.Decimal
			err := conn.QueryRow(ctx, "SELECT $1::numeric[]", inputSlice).Scan(&outputSlice)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Benchmark memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("Decimal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var d pgxdecimal.Decimal
			numeric := benchmarkNumericValues[i%len(benchmarkNumericValues)]
			d.ScanNumeric(numeric)
		}
	})

	b.Run("NullDecimal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var d pgxdecimal.NullDecimal
			numeric := benchmarkNumericValues[i%len(benchmarkNumericValues)]
			d.ScanNumeric(numeric)
		}
	})
}

// Benchmark error handling paths
func BenchmarkErrorHandling(b *testing.B) {
	b.Run("ScanNaN", func(b *testing.B) {
		var d pgxdecimal.Decimal
		nanNumeric := pgtype.Numeric{Valid: true, NaN: true}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d.ScanNumeric(nanNumeric)
		}
	})

	b.Run("ScanNull", func(b *testing.B) {
		var d pgxdecimal.Decimal
		nullNumeric := pgtype.Numeric{Valid: false}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d.ScanNumeric(nullNumeric)
		}
	})

	b.Run("ScanInfinity", func(b *testing.B) {
		var d pgxdecimal.Decimal
		infNumeric := pgtype.Numeric{Valid: true, InfinityModifier: pgtype.Infinity}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d.ScanNumeric(infNumeric)
		}
	})
}
