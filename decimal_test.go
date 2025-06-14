package decimal_test

import (
	"context"
	"math"
	"os"
	"strings"
	"testing"

	"log"

	"github.com/ingothierack/decimal128"
	pgxdecimal "github.com/ingothierack/pgx-woodsbury-decimal128"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxtest"
	"github.com/stretchr/testify/require"
)

var defaultConnTestRunner pgxtest.ConnTestRunner

func init() {
	defaultConnTestRunner = pgxtest.DefaultConnTestRunner()
	defaultConnTestRunner.AfterConnect = func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		pgxdecimal.Register(conn.TypeMap())
	}
	defaultConnTestRunner.CreateConfig = func(ctx context.Context, t testing.TB) *pgx.ConnConfig {
		config, err := pgx.ParseConfig(os.Getenv("PGX_TEST_DATABASE"))
		require.NoError(t, err)
		return config
	}
	log.Println("init complete")
}

func TestCodecDecodeValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		original := decimal128.MustParse("1.2345678901234")

		rows, err := conn.Query(context.Background(), `select $1::numeric`, original)
		require.NoError(t, err)

		for rows.Next() {
			values, err := rows.Values()
			require.NoError(t, err)

			require.Len(t, values, 1)
			v0, ok := values[0].(decimal128.Decimal)
			require.True(t, ok)
			require.Equal(t, original, v0)
		}

		require.NoError(t, rows.Err())

		rows, err = conn.Query(context.Background(), `select $1::numeric`, nil)
		require.NoError(t, err)

		for rows.Next() {
			values, err := rows.Values()
			require.NoError(t, err)

			require.Len(t, values, 1)
			require.Equal(t, nil, values[0])
		}

		require.NoError(t, rows.Err())
	})
}

func TestNaN(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		var d decimal128.Decimal
		err := conn.QueryRow(context.Background(), `select 'NaN'::numeric`).Scan(&d)
		require.EqualError(t, err, `can't scan into dest[0] (col: numeric): cannot scan NaN into *decimal128.Decimal`)
	})
}

func TestArray(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		inputSlice := []decimal128.Decimal{}

		for i := range 10 {
			d := decimal128.FromInt64(int64(i))
			inputSlice = append(inputSlice, d)
		}

		var outputSlice []decimal128.Decimal
		err := conn.QueryRow(context.Background(), `select $1::numeric[]`, inputSlice).Scan(&outputSlice)
		require.NoError(t, err)

		require.Equal(t, len(inputSlice), len(outputSlice))
		for i := range len(inputSlice) {
			require.True(t, outputSlice[i].Equal(inputSlice[i]))
		}
	})
}

func isExpectedEqDecimal(a decimal128.Decimal) func(any) bool {
	return func(v any) bool {
		return a.Equal(v.(decimal128.Decimal))
	}
}

func isExpectedEqString(a string) func(any) bool {
	return func(v any) bool {
		val := v.(decimal128.Decimal).String()
		ret := strings.Compare(a, val)
		if ret == 0 {
			return true
		}
		return false
	}
}

func isExpectedEqNullDecimal(a pgxdecimal.NullDecimal) func(any) bool {
	return func(v any) bool {
		b := v.(pgxdecimal.NullDecimal)
		return a.Valid == b.Valid && a.Decimal.Equal(b.Decimal)
	}
}

func TestValueRoundTrip(t *testing.T) {
	pgxtest.RunValueRoundTripTests(context.Background(), t, defaultConnTestRunner, nil, "numeric", []pgxtest.ValueRoundTripTest{
		{
			Param:  decimal128.MustParse("1"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("1")),
		},
		{
			Param:  decimal128.MustParse("0.000012345"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("0.000012345")),
		},
		{
			Param:  decimal128.MustParse("123456.123456"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("123456.123456")),
		},
		{
			Param:  decimal128.MustParse("-1"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("-1")),
		},
		{
			Param:  decimal128.MustParse("-0.000012345"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("-0.000012345")),
		},
		{
			Param:  decimal128.MustParse("-123456.123456"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("-123456.123456")),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("1"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("1"), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("0.000012345"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("0.000012345"), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("123456.123456"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("123456.123456"), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-1"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-1"), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-0.000012345"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-0.000012345"), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-123456.123456"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-123456.123456"), Valid: true}),
		},
	})
}

func TestValueRoundTripNumeric(t *testing.T) {
	pgxtest.RunValueRoundTripTests(context.Background(), t, defaultConnTestRunner, nil, "numeric", []pgxtest.ValueRoundTripTest{
		{
			Param:  decimal128.MustParse("1"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("1")),
		},
		{
			Param:  decimal128.MustParse("123456789012345.123456789012"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("123456789012345.123456789012")),
		},
		{
			Param:  decimal128.MustParse("12345678901234567890.12345678901234567890"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("12345678901234567890.12345678901234567890")),
		},
		// {
		// 	Param:  decimal128.MustParse("1234567890123456789012345678901234567890.1234567890123456789012345678901234567890"),
		// 	Result: new(decimal128.Decimal),
		// 	Test:   isExpectedEqString("1.2345678901234567890123456789012345678901234567890123456789012345678901234567890e+39"),
		// },
		{
			Param:  decimal128.MustParse("9_345_678_901_234_567_890.123_456_789_012_345"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqString("9.345678901234567890123456789012345e+18"),
		},
		{
			Param:  decimal128.MustParse("-9_345_678_901_234_567_890.123_456_789_012_345"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqString("-9.345678901234567890123456789012345e+18"),
		},
	})
}

func TestValueRoundTripFloat8(t *testing.T) {
	pgxtest.RunValueRoundTripTests(context.Background(), t, defaultConnTestRunner, nil, "float8", []pgxtest.ValueRoundTripTest{
		{
			Param:  decimal128.MustParse("1"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("1")),
		},
		{
			Param:  decimal128.MustParse("0.000012345"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("0.000012345")),
		},
		{
			Param:  decimal128.MustParse("123456.123456"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("123456.123456")),
		},
		{
			Param:  decimal128.MustParse("-1"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("-1")),
		},
		{
			Param:  decimal128.MustParse("-0.000012345"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("-0.000012345")),
		},
		{
			Param:  decimal128.MustParse("-123456.123456"),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.MustParse("-123456.123456")),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("1"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("1"), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("0.000012345"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("0.000012345"), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("123456.123456"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("123456.123456"), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-1"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-1"), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-0.000012345"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-0.000012345"), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-123456.123456"), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.MustParse("-123456.123456"), Valid: true}),
		},
	})
}

func TestValueRoundTripInt8(t *testing.T) {
	pgxtest.RunValueRoundTripTests(context.Background(), t, defaultConnTestRunner, nil, "int8", []pgxtest.ValueRoundTripTest{
		{
			Param:  decimal128.FromInt64(0),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.FromInt64(0)),
		},
		{
			Param:  decimal128.FromInt64(1),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.FromInt64(1)),
		},
		{
			Param:  decimal128.FromInt64(math.MaxInt64),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.FromInt64(math.MaxInt64)),
		},
		{
			Param:  decimal128.FromInt64(-1),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.FromInt64(-1)),
		},
		{
			Param:  decimal128.FromInt64(math.MinInt64),
			Result: new(decimal128.Decimal),
			Test:   isExpectedEqDecimal(decimal128.FromInt64(math.MinInt64)),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.FromInt64(0), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.FromInt64(0), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.FromInt64(1), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.FromInt64(1), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.FromInt64(math.MaxInt64), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.FromInt64(math.MaxInt64), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.FromInt64(-1), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.FromInt64(-1), Valid: true}),
		},
		{
			Param:  pgxdecimal.NullDecimal{Decimal: decimal128.FromInt64(math.MinInt64), Valid: true},
			Result: new(pgxdecimal.NullDecimal),
			Test:   isExpectedEqNullDecimal(pgxdecimal.NullDecimal{Decimal: decimal128.FromInt64(math.MinInt64), Valid: true}),
		},
	})
}
