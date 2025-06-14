# Benchmarks for pgx-woodsbury-decimal128

This document provides comprehensive benchmark results and analysis for the pgx-woodsbury-decimal128 package, which provides PostgreSQL decimal128 integration for the pgx driver.

## Running Benchmarks

To run all benchmarks:
```bash
go test -bench=. -benchmem -run=^$
```

To run specific benchmarks:
```bash
go test -bench=BenchmarkDecimalScanNumeric -benchmem -run=^$
```

To run benchmarks without database operations (short mode):
```bash
go test -bench=. -benchmem -run=^$ -short
```

## Benchmark Results

Results from Apple M2 Pro (12 cores):

### Core Operations

| Benchmark | ns/op | B/op | allocs/op | Description |
|-----------|-------|------|-----------|-------------|
| BenchmarkDecimalScanNumeric | 31.81 | 7 | 0 | Scanning pgtype.Numeric to Decimal |
| BenchmarkDecimalNumericValue | 51.01 | 57 | 2 | Converting Decimal to pgtype.Numeric |
| BenchmarkDecimalScanFloat64 | 151.4 | 133 | 2 | Scanning pgtype.Float8 to Decimal |
| BenchmarkDecimalFloat64Value | 191.2 | 0 | 0 | Converting Decimal to pgtype.Float8 |
| BenchmarkDecimalScanInt64 | 1.870 | 0 | 0 | Scanning pgtype.Int8 to Decimal |
| BenchmarkDecimalInt64Value | 45.83 | 0 | 0 | Converting Decimal to pgtype.Int8 |

### NullDecimal Operations

| Benchmark | ns/op | B/op | allocs/op | Description |
|-----------|-------|------|-----------|-------------|
| BenchmarkNullDecimalScanNumeric | 31.81 | 7 | 0 | Scanning pgtype.Numeric to NullDecimal |
| BenchmarkNullDecimalNumericValue | 60.29 | 61 | 2 | Converting NullDecimal to pgtype.Numeric |
| BenchmarkNullDecimalScanFloat64 | 151.1 | 133 | 2 | Scanning pgtype.Float8 to NullDecimal |
| BenchmarkNullDecimalFloat64Value | 177.5 | 0 | 0 | Converting NullDecimal to pgtype.Float8 |
| BenchmarkNullDecimalScanInt64 | 1.873 | 0 | 0 | Scanning pgtype.Int8 to NullDecimal |
| BenchmarkNullDecimalInt64Value | 78.44 | 102 | 4 | Converting NullDecimal to pgtype.Int8 |

### Plan Wrappers

| Benchmark | ns/op | B/op | allocs/op | Description |
|-----------|-------|------|-----------|-------------|
| BenchmarkTryWrapNumericEncodePlan | 1.184 | 0 | 0 | Type detection for encoding |
| BenchmarkTryWrapNumericScanPlan | 1.027 | 0 | 0 | Type detection for scanning |

### Error Handling

| Benchmark | ns/op | B/op | allocs/op | Description |
|-----------|-------|------|-----------|-------------|
| BenchmarkErrorHandling/ScanNaN | 1.941 | 0 | 0 | Handling NaN values |
| BenchmarkErrorHandling/ScanNull | 1.938 | 0 | 0 | Handling NULL values |
| BenchmarkErrorHandling/ScanInfinity | 87.61 | 64 | 2 | Handling infinity values |

## Performance Analysis

### Key Insights

1. **Int64 Operations are Fastest**: Scanning from `pgtype.Int8` is extremely fast (< 2 ns/op) with zero allocations.

2. **Numeric Operations are Efficient**: Both Decimal and NullDecimal show similar performance for numeric operations (~32 ns/op for scanning).

3. **Float64 Operations are Costlier**: Float64 conversions require string parsing, resulting in higher latency (~150 ns/op) and allocations.

4. **Error Handling is Fast**: Error paths for NULL and NaN values are extremely efficient (< 2 ns/op).

5. **Plan Wrappers are Negligible**: Type detection overhead is minimal (< 2 ns/op).

### Memory Allocation Patterns

- **Zero-allocation operations**: Int64 scanning, Float64 value conversion, error handling
- **Minimal allocations**: Numeric operations (7 B/op, 0-2 allocs/op)
- **Higher allocations**: Float64 scanning (133 B/op, 2 allocs/op) due to string conversion

### Recommendations

1. **Prefer Int64 when possible**: If your decimal values fit in int64 range, use int64 operations for maximum performance.

2. **Use Numeric for precision**: For decimal precision requirements, the numeric operations provide good performance with minimal allocations.

3. **Avoid Float64 for high-throughput**: Float64 conversions have higher overhead due to string parsing.

4. **NullDecimal overhead is minimal**: The nullable version adds very little overhead for most operations.

## Database Benchmarks

Database benchmarks require a PostgreSQL connection and are skipped in short mode. Set the `PGX_TEST_DATABASE` environment variable to run these benchmarks:

```bash
export PGX_TEST_DATABASE="postgres://user:password@localhost/testdb"
go test -bench=BenchmarkDatabase -benchmem -run=^$
```

### Database Operations

- **BenchmarkDatabaseRoundTrip**: Tests full database round-trip performance
- **BenchmarkArrayOperations**: Tests array handling performance (100 elements)

## Comparison with Other Libraries

The benchmarks can be extended to compare with other decimal libraries:

```go
// Example comparison benchmark
func BenchmarkComparison(b *testing.B) {
    b.Run("pgx-woodsbury-decimal128", func(b *testing.B) {
        // Current implementation
    })

    b.Run("shopspring-decimal", func(b *testing.B) {
        // Alternative implementation
    })
}
```

## Benchmark Environment

- **Platform**: Apple M2 Pro
- **Cores**: 12
- **Go Version**: 1.24+
- **Architecture**: arm64
- **OS**: macOS

Results may vary on different platforms and Go versions.

## Benchmark Summary

The comprehensive benchmark suite for pgx-woodsbury-decimal128 demonstrates excellent performance characteristics:

### Performance Highlights

- **Ultra-fast Int64 operations**: < 2 ns/op with zero allocations
- **Efficient Numeric operations**: ~32 ns/op with minimal memory usage (7 B/op)
- **Robust error handling**: < 2 ns/op for common error cases
- **Minimal plan overhead**: Type detection adds < 2 ns/op

### Memory Efficiency

The package shows excellent memory efficiency with most operations requiring minimal allocations:
- Int64 operations: 0 allocations
- Numeric operations: 0-2 allocations with 7-57 bytes
- Error paths: 0 allocations

### Usage Examples

#### Basic Benchmarking

```go
// Run all benchmarks
go test -bench=. -benchmem -run=^$

// Run specific operation benchmarks
go test -bench=BenchmarkDecimalScanNumeric -benchmem -run=^$

// Run with multiple iterations for statistical significance
go test -bench=BenchmarkDecimal -benchmem -run=^$ -count=5
```

#### Custom Benchmark Integration

```go
func BenchmarkYourOperation(b *testing.B) {
    var d pgxdecimal.Decimal
    numeric := pgtype.Numeric{
        Int: big.NewInt(12345),
        Exp: -2,
        Valid: true,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        d.ScanNumeric(numeric)
    }
}
```

#### Performance Testing in Applications

```go
func TestApplicationPerformance(t *testing.T) {
    // Test with realistic data volumes
    values := make([]decimal128.Decimal, 10000)
    for i := range values {
        values[i] = decimal128.FromInt64(int64(i))
    }

    start := time.Now()
    for _, v := range values {
        var numeric pgtype.Numeric
        d := pgxdecimal.Decimal(v)
        d.NumericValue()
    }
    duration := time.Since(start)

    t.Logf("Processed %d values in %v (%.2f ns/op)",
           len(values), duration, float64(duration.Nanoseconds())/float64(len(values)))
}
```

## Optimization Tips

1. **Choose the right type for your use case**:
   - Use `int64` operations when values fit in int64 range
   - Use `numeric` operations when precision is critical
   - Avoid `float64` operations in high-throughput scenarios

2. **Batch operations when possible**:
   - Array operations are more efficient than individual operations
   - Consider batching database queries

3. **Pre-allocate buffers**:
   - The package uses sync.Pool for buffer management
   - Consider similar patterns in your application code

4. **Monitor allocations**:
   - Use `-benchmem` flag to track memory allocations
   - Profile your application to identify allocation hotspots
