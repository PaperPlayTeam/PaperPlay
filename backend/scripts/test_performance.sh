#!/bin/bash

# PaperPlay Performance Test Suite
# Tests API endpoints for performance benchmarks

set -e

# Configuration
BASE_URL="http://localhost:8080"
CONCURRENT_USERS=10
REQUESTS_PER_USER=5
TIMEOUT=30s

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}=================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=================================${NC}"
}

print_test() {
    echo -e "${YELLOW}Testing: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ PASS: $1${NC}"
}

print_error() {
    echo -e "${RED}✗ FAIL: $1${NC}"
}

# Test endpoint response time
test_response_time() {
    local endpoint="$1"
    local method="${2:-GET}"
    local threshold="${3:-500}"
    local test_name="$4"
    
    print_test "$test_name"
    
    start_time=$(date +%s%N)
    if [ "$method" = "GET" ]; then
        curl -s -o /dev/null "$BASE_URL$endpoint"
    else
        curl -s -o /dev/null -X "$method" "$BASE_URL$endpoint"
    fi
    end_time=$(date +%s%N)
    
    response_time=$(((end_time - start_time) / 1000000))
    
    if [ $response_time -lt $threshold ]; then
        print_success "$test_name: ${response_time}ms (< ${threshold}ms)"
    else
        print_error "$test_name: ${response_time}ms (>= ${threshold}ms)"
    fi
}

# Test concurrent requests
test_concurrent_requests() {
    local endpoint="$1"
    local concurrent_users="$2"
    local test_name="$3"
    
    print_test "$test_name"
    
    start_time=$(date +%s%N)
    
    for ((i=1; i<=concurrent_users; i++)); do
        curl -s -o /dev/null "$BASE_URL$endpoint" &
    done
    
    wait
    end_time=$(date +%s%N)
    
    total_time=$(((end_time - start_time) / 1000000))
    avg_time=$((total_time / concurrent_users))
    
    if [ $total_time -lt 2000 ]; then
        print_success "$test_name: ${concurrent_users} requests in ${total_time}ms (avg: ${avg_time}ms)"
    else
        print_error "$test_name: ${concurrent_users} requests in ${total_time}ms (too slow)"
    fi
}

# Test throughput
test_throughput() {
    local endpoint="$1"
    local duration="$2"
    local test_name="$3"
    
    print_test "$test_name"
    
    request_count=0
    start_time=$(date +%s)
    end_time=$((start_time + duration))
    
    while [ $(date +%s) -lt $end_time ]; do
        curl -s -o /dev/null "$BASE_URL$endpoint" &
        request_count=$((request_count + 1))
        
        # Limit concurrent requests
        if [ $((request_count % 10)) -eq 0 ]; then
            wait
        fi
    done
    
    wait
    
    throughput=$((request_count / duration))
    
    if [ $throughput -gt 50 ]; then
        print_success "$test_name: ${throughput} requests/second"
    else
        print_error "$test_name: ${throughput} requests/second (too low)"
    fi
}

# Test memory usage during load
test_memory_usage() {
    local endpoint="$1"
    local test_name="$2"
    
    print_test "$test_name"
    
    # Get initial memory usage
    initial_memory=$(curl -s "$BASE_URL/metrics" | grep "paperplay_memory_usage_bytes" | awk '{print $2}' || echo "0")
    
    # Generate load
    for i in {1..20}; do
        curl -s -o /dev/null "$BASE_URL$endpoint" &
    done
    wait
    
    # Get final memory usage
    final_memory=$(curl -s "$BASE_URL/metrics" | grep "paperplay_memory_usage_bytes" | awk '{print $2}' || echo "0")
    
    memory_increase=$((final_memory - initial_memory))
    
    if [ $memory_increase -lt 10485760 ]; then  # 10MB
        print_success "$test_name: Memory increase ${memory_increase} bytes (< 10MB)"
    else
        print_error "$test_name: Memory increase ${memory_increase} bytes (>= 10MB)"
    fi
}

# Main performance tests
main() {
    print_header "PAPERPLAY PERFORMANCE TESTS"
    
    echo -e "${BLUE}Base URL: $BASE_URL${NC}"
    echo -e "${BLUE}Concurrent Users: $CONCURRENT_USERS${NC}"
    echo -e "${BLUE}Requests per User: $REQUESTS_PER_USER${NC}"
    echo ""
    
    # Check if server is running
    if ! curl -s "$BASE_URL/health" > /dev/null; then
        echo -e "${RED}Server is not running. Please start the server first.${NC}"
        exit 1
    fi
    
    print_header "RESPONSE TIME TESTS"
    test_response_time "/health" "GET" 100 "Health endpoint response time"
    test_response_time "/metrics" "GET" 200 "Metrics endpoint response time"
    
    print_header "CONCURRENT REQUEST TESTS"
    test_concurrent_requests "/health" 5 "5 concurrent health requests"
    test_concurrent_requests "/health" 10 "10 concurrent health requests"
    test_concurrent_requests "/metrics" 5 "5 concurrent metrics requests"
    
    print_header "THROUGHPUT TESTS"
    test_throughput "/health" 5 "Health endpoint throughput (5 seconds)"
    
    print_header "MEMORY USAGE TESTS"
    test_memory_usage "/health" "Health endpoint memory usage"
    
    print_header "LOAD TESTING"
    print_test "Sustained load test"
    
    # Sustained load test
    echo "Generating sustained load for 10 seconds..."
    for i in {1..10}; do
        for j in {1..5}; do
            curl -s -o /dev/null "$BASE_URL/health" &
        done
        sleep 1
    done
    wait
    
    # Check if server is still responsive
    if curl -s "$BASE_URL/health" > /dev/null; then
        print_success "Server remained responsive during sustained load"
    else
        print_error "Server became unresponsive during sustained load"
    fi
    
    echo ""
    echo -e "${GREEN}Performance tests completed!${NC}"
}

# Run main function
main "$@" 