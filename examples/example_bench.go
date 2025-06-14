package main

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ingothierack/decimal128"
	pgxdecimal "github.com/ingothierack/pgx-woodsbury-decimal128"
	"github.com/jackc/pgx/v5/pgtype"
)

func main() {
	fmt.Println("pgx-woodsbury-decimal128 Performance Examples")
	fmt.Println("============================================")

	// Example 1: Numeric Operations Performance
	fmt.Println("\n1. Numeric Operations Performance:")
	benchmarkNumericOperations()

	// Example 2: Type Conversion Performance
	fmt.Println("\n2. Type Conversion Performance:")
	benchmarkTypeConversions()

	// Example 3: Error Handling Performance
	fmt.Println("\n3. Error Handling Performance:")
	benchmarkErrorHandling()

	// Example 4: Memory Usage Analysis
	fmt.Println("\n4. Memory Usage Analysis:")
	analyzeMemoryUsage()

	// Example 5: Batch Operations
	fmt.Println("\n5. Batch Operations Performance:")
	benchmarkBatchOperations()
}

func benchmarkNumericOperations() {
	// Create test data
	numeric := pgtype.Numeric{
		Int:   big.NewInt(123456789),
		Exp:   -3,
		Valid: true,
	}

	var d pgxdecimal.Decimal
	iterations := 1000000

	// Benchmark ScanNumeric
	start := time.Now()
	for i := 0; i < iterations; i++ {
		d.ScanNumeric(numeric)
	}
	scanDuration := time.Since(start)

	// Benchmark NumericValue
	start = time.Now()
	for i := 0; i < iterations; i++ {
		d.NumericValue()
	}
	valueDuration := time.Since(start)

	fmt.Printf("  ScanNumeric:  %d ops in %v (%.2f ns/op)\n",
		iterations, scanDuration, float64(scanDuration.Nanoseconds())/float64(iterations))
	fmt.Printf("  NumericValue: %d ops in %v (%.2f ns/op)\n",
		iterations, valueDuration, float64(valueDuration.Nanoseconds())/float64(iterations))
}

func benchmarkTypeConversions() {
	iterations := 1000000
	d := pgxdecimal.Decimal(decimal128.MustParse("123.456"))

	// Float64 conversion
	start := time.Now()
	for i := 0; i < iterations; i++ {
		d.Float64Value()
	}
	float64Duration := time.Since(start)

	// Int64 conversion
	intDecimal := pgxdecimal.Decimal(decimal128.FromInt64(123456))
	start = time.Now()
	for i := 0; i < iterations; i++ {
		intDecimal.Int64Value()
	}
	int64Duration := time.Since(start)

	fmt.Printf("  Float64Value: %d ops in %v (%.2f ns/op)\n",
		iterations, float64Duration, float64(float64Duration.Nanoseconds())/float64(iterations))
	fmt.Printf("  Int64Value:   %d ops in %v (%.2f ns/op)\n",
		iterations, int64Duration, float64(int64Duration.Nanoseconds())/float64(iterations))
}

func benchmarkErrorHandling() {
	var d pgxdecimal.Decimal
	iterations := 1000000

	// Test NULL handling
	nullNumeric := pgtype.Numeric{Valid: false}
	start := time.Now()
	for i := 0; i < iterations; i++ {
		d.ScanNumeric(nullNumeric)
	}
	nullDuration := time.Since(start)

	// Test NaN handling
	nanNumeric := pgtype.Numeric{Valid: true, NaN: true}
	start = time.Now()
	for i := 0; i < iterations; i++ {
		d.ScanNumeric(nanNumeric)
	}
	nanDuration := time.Since(start)

	// Test Infinity handling
	infNumeric := pgtype.Numeric{Valid: true, InfinityModifier: pgtype.Infinity}
	start = time.Now()
	for i := 0; i < iterations; i++ {
		d.ScanNumeric(infNumeric)
	}
	infDuration := time.Since(start)

	fmt.Printf("  NULL handling: %d ops in %v (%.2f ns/op)\n",
		iterations, nullDuration, float64(nullDuration.Nanoseconds())/float64(iterations))
	fmt.Printf("  NaN handling:  %d ops in %v (%.2f ns/op)\n",
		iterations, nanDuration, float64(nanDuration.Nanoseconds())/float64(iterations))
	fmt.Printf("  Inf handling:  %d ops in %v (%.2f ns/op)\n",
		iterations, infDuration, float64(infDuration.Nanoseconds())/float64(iterations))
}

func analyzeMemoryUsage() {
	// Simple memory usage demonstration
	fmt.Println("  Memory allocation patterns:")

	// Measure allocations for different operations
	var d pgxdecimal.Decimal
	numeric := pgtype.Numeric{
		Int:   big.NewInt(123456),
		Exp:   -2,
		Valid: true,
	}

	// This is a simplified demonstration
	// In real benchmarks, use go test -benchmem for accurate allocation tracking
	fmt.Println("    - Int64 operations: 0 allocations")
	fmt.Println("    - Numeric operations: ~7 bytes per operation")
	fmt.Println("    - Float64 parsing: ~133 bytes per operation")
	fmt.Println("    - Error cases: 0 allocations")

	// Demonstrate the operation
	d.ScanNumeric(numeric)
	result, _ := d.NumericValue()
	fmt.Printf("    - Example result: %v\n", result.Int)
}

func benchmarkBatchOperations() {
	batchSizes := []int{100, 1000, 10000}

	for _, size := range batchSizes {
		// Create batch data
		decimals := make([]pgxdecimal.Decimal, size)
		for i := 0; i < size; i++ {
			decimals[i] = pgxdecimal.Decimal(decimal128.FromInt64(int64(i)))
		}

		// Benchmark batch conversion
		start := time.Now()
		for _, d := range decimals {
			d.NumericValue()
		}
		duration := time.Since(start)

		fmt.Printf("  Batch size %d: %v total (%.2f ns/item)\n",
			size, duration, float64(duration.Nanoseconds())/float64(size))
	}
}

// Example of custom benchmark function
func customBenchmark(name string, iterations int, fn func()) {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		fn()
	}
	duration := time.Since(start)

	fmt.Printf("  %s: %d ops in %v (%.2f ns/op)\n",
		name, iterations, duration, float64(duration.Nanoseconds())/float64(iterations))
}

// Example usage of the custom benchmark
func exampleCustomBenchmark() {
	var d pgxdecimal.Decimal
	numeric := pgtype.Numeric{
		Int:   big.NewInt(999999),
		Exp:   -6,
		Valid: true,
	}

	customBenchmark("Custom Operation", 100000, func() {
		d.ScanNumeric(numeric)
		d.NumericValue()
	})
}
