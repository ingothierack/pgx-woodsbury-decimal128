package decimal

import (
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/woodsbury/decimal128"
)

func TestConvertDecimal128(t *testing.T) {
	test_int64 := "-978901234567890123456789"
	var test_pgnumeric *pgtype.Numeric
	var test_bigint big.Int

	test_bigint.SetString(test_int64, 10)

	test, _ := new(big.Int).SetString("978901234567890123456789", 10)

	test_pgnumeric = &pgtype.Numeric{
		Valid: true,
		Int:   test,
		Exp:   -9,
	}

	var dec decimal128.Decimal
	dec, _ = decimal128.Parse("-978_901_234_567_890.123_456_789")

	var dec1 decimal128.Decimal
	dec1, _ = decimal128.Parse("978_901_234_567_890.123_456_789")

	var dec2 decimal128.Decimal
	dec2.Compose(0, false, test_bigint.Bytes(), test_pgnumeric.Exp)

	var dec3 decimal128.Decimal
	testneg := false
	if test_pgnumeric.Int.Sign() < 0 {
		testneg = true
	}
	dec3.Compose(0, testneg, test_pgnumeric.Int.Bytes(), test_pgnumeric.Exp)

	var intvar *big.Int
	var exp int32

	intvar = dec3.Int(intvar)
	testdec := decimal128.Exp(dec3)

	temp1, _ := test_pgnumeric.Float64Value()
	t.Logf("pgtype numeric: %v\n", test_pgnumeric)
	t.Logf("pgtype numeric: %v\n", temp1)
	t.Logf("decimal dec: %v\n", dec.String())
	t.Logf("decimal dec: %-20.10f\n", dec)
	t.Logf("decimal dec1: %v\n", dec1.String())
	t.Logf("decimal dec: %-20.10f\n", dec1)
	t.Logf("decimal dec2: %v\n", dec2.String())
	t.Logf("decimal dec: %-20.10f\n", dec2)
	t.Logf("decimal dec3: %-20.10f\n", dec3)
	t.Logf("decimal testdec: %v\n", testdec)

	t.Logf("bigintval: %v, exp: %v", intvar, exp)

}
