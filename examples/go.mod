module examples

go 1.24.0

toolchain go1.24.4

require (
	github.com/ingothierack/decimal128 v0.0.0-20250423052917-88712c3525f8
	github.com/ingothierack/pgx-woodsbury-decimal128 v0.0.0
	github.com/jackc/pgx/v5 v5.7.5
)

replace github.com/ingothierack/pgx-woodsbury-decimal128 => ../
