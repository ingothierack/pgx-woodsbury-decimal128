#!/bin/bash

# pgx-woodsbury-decimal128 Benchmark Runner
# =========================================
# This script runs comprehensive benchmarks for the pgx-woodsbury-decimal128 package

set -e

echo "pgx-woodsbury-decimal128 Benchmark Suite"
echo "========================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_header() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_warning() {
    echo -e "${YELLOW}$1${NC}"
}

print_error() {
    echo -e "${RED}$1${NC}"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "decimal.go" ] || [ ! -f "decimal_test.go" ]; then
    print_error "Please run this script from the pgx-woodsbury-decimal128 directory"
    exit 1
fi

# Default options
RUN_ALL=true
RUN_QUICK=false
RUN_DATABASE=false
RUN_MEMORY=false
COUNT=1
BENCHTIME=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --quick|-q)
            RUN_QUICK=true
            RUN_ALL=false
            shift
            ;;
        --database|-d)
            RUN_DATABASE=true
            RUN_ALL=false
            shift
            ;;
        --memory|-m)
            RUN_MEMORY=true
            RUN_ALL=false
            shift
            ;;
        --count|-c)
            COUNT="$2"
            shift 2
            ;;
        --benchtime|-t)
            BENCHTIME="-benchtime=$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --quick, -q      Run quick benchmarks only"
            echo "  --database, -d   Run database benchmarks only"
            echo "  --memory, -m     Run memory allocation benchmarks only"
            echo "  --count, -c N    Run benchmarks N times (default: 1)"
            echo "  --benchtime, -t T Set benchmark time (e.g., 10s, 100000x)"
            echo "  --help, -h       Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                    # Run all benchmarks"
            echo "  $0 --quick           # Run quick benchmarks"
            echo "  $0 --count 5         # Run benchmarks 5 times"
            echo "  $0 --benchtime 10s   # Run each benchmark for 10 seconds"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Build benchmark command
BENCH_CMD="go test -bench="
BENCH_FLAGS="-benchmem -run=^$ -v"

if [ "$COUNT" -gt 1 ]; then
    BENCH_FLAGS="$BENCH_FLAGS -count=$COUNT"
fi

if [ -n "$BENCHTIME" ]; then
    BENCH_FLAGS="$BENCH_FLAGS $BENCHTIME"
fi

print_header "Environment Information"
echo "Go Version: $(go version)"
echo "OS: $(uname -s)"
echo "Architecture: $(uname -m)"
echo "CPU Cores: $(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 'unknown')"
echo ""

# Run benchmarks based on options
if [ "$RUN_ALL" = true ] || [ "$RUN_QUICK" = true ]; then
    print_header "Running Core Operation Benchmarks"
    ${BENCH_CMD}BenchmarkDecimal ${BENCH_FLAGS}
    echo ""
fi

if [ "$RUN_ALL" = true ]; then
    print_header "Running NullDecimal Benchmarks"
    ${BENCH_CMD}BenchmarkNullDecimal ${BENCH_FLAGS}
    echo ""

    print_header "Running Plan Wrapper Benchmarks"
    ${BENCH_CMD}BenchmarkTryWrap ${BENCH_FLAGS}
    echo ""

    print_header "Running Error Handling Benchmarks"
    ${BENCH_CMD}BenchmarkErrorHandling ${BENCH_FLAGS}
    echo ""

    print_header "Running Composition Benchmarks"
    ${BENCH_CMD}BenchmarkDecimalComposition ${BENCH_FLAGS}
    echo ""
fi

if [ "$RUN_ALL" = true ] || [ "$RUN_MEMORY" = true ]; then
    print_header "Running Memory Allocation Benchmarks"
    ${BENCH_CMD}BenchmarkMemoryAllocation ${BENCH_FLAGS}
    echo ""
fi

if [ "$RUN_DATABASE" = true ]; then
    print_header "Running Database Benchmarks"
    if [ -z "$PGX_TEST_DATABASE" ]; then
        print_warning "PGX_TEST_DATABASE environment variable not set"
        print_warning "Database benchmarks will be skipped"
        print_warning "Set PGX_TEST_DATABASE to run database benchmarks:"
        print_warning "export PGX_TEST_DATABASE='postgres://user:password@localhost/testdb'"
    else
        print_success "Running database benchmarks with: $PGX_TEST_DATABASE"
        ${BENCH_CMD}BenchmarkDatabase ${BENCH_FLAGS}
    fi
    echo ""
fi

# Run example program if it exists
if [ -f "examples/example_bench.go" ]; then
    print_header "Running Performance Examples"
    if (cd examples && go run example_bench.go); then
        print_success "Performance examples completed successfully"
    else
        print_warning "Performance examples failed or had warnings"
    fi
    echo ""
fi

print_header "Benchmark Analysis"
echo "For detailed analysis, consider using:"
echo "  go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof"
echo "  go tool pprof cpu.prof"
echo "  go tool pprof mem.prof"
echo ""

print_header "Performance Tips"
echo "1. Use Int64 operations when values fit in int64 range"
echo "2. Prefer Numeric operations for high precision requirements"
echo "3. Avoid Float64 operations in high-throughput scenarios"
echo "4. Consider batching database operations"
echo "5. Monitor memory allocations in production"
echo ""

print_success "Benchmark suite completed!"

# Generate benchmark report if benchstat is available
if command -v benchstat &> /dev/null && [ "$COUNT" -gt 1 ]; then
    print_header "Statistical Analysis Available"
    echo "Run 'benchstat' on the output files for statistical analysis"
    echo "Example: go test -bench=. -count=10 > bench.txt && benchstat bench.txt"
fi
