# pgx-woodsbury-decimal128

A high-performance PostgreSQL decimal128 integration for the [pgx](https://github.com/jackc/pgx) driver, providing seamless conversion between PostgreSQL numeric types and Go's decimal128.Decimal values.

## Features

- **Fast and efficient**: Optimized for high-performance database operations
- **Type-safe**: Full integration with pgx's type system
- **Zero-allocation paths**: Critical operations avoid memory allocations
- **Comprehensive support**: Handles numeric, float8, and int8 PostgreSQL types
- **Null handling**: Built-in support for nullable decimals
- **Array support**: Full support for PostgreSQL arrays
- **Error resilient**: Proper handling of NaN, infinity, and null values

## Installation

```bash
go get github.com/ingothierack/pgx-woodsbury-decimal128
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/ingothierack/decimal128"
    pgxdecimal "github.com/ingothierack/pgx-woodsbury-decimal128"
    "github.com/jackc/pgx/v5"
)

func main() {
    conn, err := pgx.Connect(context.Background(), "postgres://...")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close(context.Background())

    // Register the decimal types
    pgxdecimal.Register(conn.TypeMap())

    // Use decimal128.Decimal directly
    var result decimal128.Decimal
    err = conn.QueryRow(context.Background(),
        "SELECT $1::numeric", decimal128.MustParse("123.456")).Scan(&result)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Result: %s", result.String())
}
```

## Supported Types

### Decimal
Non-nullable decimal type that wraps `decimal128.Decimal`:

```go
var d pgxdecimal.Decimal
err := conn.QueryRow(ctx, "SELECT 123.456::numeric").Scan(&d)
```

### NullDecimal
Nullable decimal type for handling NULL values:

```go
var nd pgxdecimal.NullDecimal
err := conn.QueryRow(ctx, "SELECT NULL::numeric").Scan(&nd)
if nd.Valid {
    log.Printf("Value: %s", nd.Decimal.String())
} else {
    log.Println("Value is NULL")
}
```

### Arrays
Full support for PostgreSQL arrays:

```go
values := []decimal128.Decimal{
    decimal128.MustParse("1.23"),
    decimal128.MustParse("4.56"),
}

var result []decimal128.Decimal
err := conn.QueryRow(ctx, "SELECT $1::numeric[]", values).Scan(&result)
```

## Performance

This library is optimized for high-performance applications:

- **Int64 operations**: < 2 ns/op with zero allocations
- **Numeric operations**: ~32 ns/op with minimal allocations
- **Error handling**: < 2 ns/op for common error cases

### Running Benchmarks

```bash
# Run all benchmarks
./run_benchmarks.sh

# Run quick benchmarks only
./run_benchmarks.sh --quick

# Run memory allocation benchmarks
./run_benchmarks.sh --memory

# Run multiple iterations for statistical analysis
./run_benchmarks.sh --count 5
```

See [BENCHMARKS.md](BENCHMARKS.md) for detailed performance analysis.

## Database Type Support

| PostgreSQL Type | Go Type | Performance |
|----------------|---------|-------------|
| `numeric` | `decimal128.Decimal` | ~32 ns/op |
| `float8` | `decimal128.Decimal` | ~150 ns/op |
| `int8` | `decimal128.Decimal` | ~2 ns/op |
| `numeric[]` | `[]decimal128.Decimal` | Batch optimized |

## Error Handling

The library properly handles PostgreSQL's special numeric values:

```go
// NaN values return an error
var d decimal128.Decimal
err := conn.QueryRow(ctx, "SELECT 'NaN'::numeric").Scan(&d)
// err: "cannot scan NaN into *decimal128.Decimal"

// Infinity values return an error
err = conn.QueryRow(ctx, "SELECT 'Infinity'::numeric").Scan(&d)
// err: "cannot scan Infinity into *decimal128.Decimal"

// NULL values work with NullDecimal
var nd NullDecimal
err = conn.QueryRow(ctx, "SELECT NULL::numeric").Scan(&nd)
// nd.Valid == false
```

## Examples

See the [examples](examples/) directory for comprehensive usage examples including:
- Basic operations
- Performance testing
- Error handling
- Batch operations
- Memory usage patterns

## Testing

Set up your test database:

```bash
export PGX_TEST_DATABASE="postgres://user:password@localhost/testdb"
go test -v
```

For database-specific tests:
```bash
export PGX_TEST_DATABASE="host=db.ipv6.m32.m31.zone database=pgtest user=pgtest password=secret"
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for your changes
4. Run the benchmark suite: `./run_benchmarks.sh`
5. Submit a pull request

## License

See [LICENSE.md](LICENSE.md) for details.
