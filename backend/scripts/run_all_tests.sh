#!/bin/bash

# PaperPlay Test Runner
# Runs all unit tests and integration tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

print_header() {
    echo -e "${BLUE}=================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}$1${NC}"
}

# Run unit tests
run_unit_tests() {
    print_header "RUNNING UNIT TESTS"
    
    cd "$PROJECT_ROOT"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go to run unit tests."
        return 1
    fi
    
    # Run unit tests with coverage
    print_info "Running Go unit tests with coverage..."
    
    if go test -v -race -coverprofile=coverage.out ./...; then
        print_success "Unit tests passed"
        
        # Generate coverage report
        if command -v go &> /dev/null; then
            coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
            print_info "Test coverage: $coverage"
            
            # Generate HTML coverage report
            go tool cover -html=coverage.out -o coverage.html
            print_info "Coverage report saved to coverage.html"
        fi
    else
        print_error "Unit tests failed"
        return 1
    fi
    
    echo ""
}

# Run integration tests
run_integration_tests() {
    print_header "RUNNING INTEGRATION TESTS"
    
    cd "$SCRIPT_DIR"
    
    # Check if integration test script exists
    if [ ! -f "test_all_endpoints.sh" ]; then
        print_error "Integration test script not found"
        return 1
    fi
    
    # Make sure script is executable
    chmod +x test_all_endpoints.sh
    
    # Run integration tests
    print_info "Running API integration tests..."
    
    if ./test_all_endpoints.sh; then
        print_success "Integration tests passed"
    else
        print_error "Integration tests failed"
        return 1
    fi
    
    echo ""
}

# Run performance tests
run_performance_tests() {
    print_header "RUNNING PERFORMANCE TESTS"
    
    cd "$SCRIPT_DIR"
    
    # Check if performance test script exists
    if [ ! -f "test_performance.sh" ]; then
        print_error "Performance test script not found"
        return 1
    fi
    
    # Make sure script is executable
    chmod +x test_performance.sh
    
    # Run performance tests
    print_info "Running performance tests..."
    
    if ./test_performance.sh; then
        print_success "Performance tests passed"
    else
        print_error "Performance tests failed"
        return 1
    fi
    
    echo ""
}

# Check test dependencies
check_dependencies() {
    print_header "CHECKING TEST DEPENDENCIES"
    
    # Check Go
    if command -v go &> /dev/null; then
        go_version=$(go version | awk '{print $3}')
        print_success "Go is installed: $go_version"
    else
        print_error "Go is not installed"
        return 1
    fi
    
    # Check curl
    if command -v curl &> /dev/null; then
        curl_version=$(curl --version | head -n1 | awk '{print $2}')
        print_success "curl is installed: $curl_version"
    else
        print_error "curl is not installed"
        return 1
    fi
    
    # Check jq
    if command -v jq &> /dev/null; then
        jq_version=$(jq --version)
        print_success "jq is installed: $jq_version"
    else
        print_error "jq is not installed (required for integration tests)"
        return 1
    fi
    
    echo ""
}

# Generate test report
generate_test_report() {
    print_header "GENERATING TEST REPORT"
    
    cd "$PROJECT_ROOT"
    
    report_file="test_report_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "PaperPlay Test Report"
        echo "Generated: $(date)"
        echo "================================="
        echo ""
        
        if [ -f "coverage.out" ]; then
            echo "Unit Test Coverage:"
            go tool cover -func=coverage.out
            echo ""
        fi
        
        echo "Test Files:"
        find . -name "*_test.go" -type f | wc -l | xargs echo "- Go unit test files:"
        find scripts/ -name "test_*.sh" -type f 2>/dev/null | wc -l | xargs echo "- Integration test scripts:"
        echo ""
        
        echo "Project Structure:"
        tree -I 'vendor|node_modules|.git|*.log' -L 3 2>/dev/null || find . -type d -name .git -prune -o -type d -print | head -20
        
    } > "$report_file"
    
    print_success "Test report generated: $report_file"
    echo ""
}

# Show usage
show_usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --unit           Run only unit tests"
    echo "  --integration    Run only integration tests"
    echo "  --performance    Run only performance tests"
    echo "  --check-deps     Check test dependencies only"
    echo "  --report         Generate test report only"
    echo "  --help           Show this help message"
    echo ""
    echo "If no options are provided, all tests will be run."
}

# Main function
main() {
    local run_unit=false
    local run_integration=false
    local run_performance=false
    local check_deps_only=false
    local generate_report_only=false
    local run_all=true
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --unit)
                run_unit=true
                run_all=false
                shift
                ;;
            --integration)
                run_integration=true
                run_all=false
                shift
                ;;
            --performance)
                run_performance=true
                run_all=false
                shift
                ;;
            --check-deps)
                check_deps_only=true
                run_all=false
                shift
                ;;
            --report)
                generate_report_only=true
                run_all=false
                shift
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    print_header "PAPERPLAY TEST SUITE"
    
    # Check dependencies first
    if ! check_dependencies; then
        print_error "Dependency check failed. Please install missing dependencies."
        exit 1
    fi
    
    if [ "$check_deps_only" = true ]; then
        exit 0
    fi
    
    if [ "$generate_report_only" = true ]; then
        generate_test_report
        exit 0
    fi
    
    local exit_code=0
    
    # Run tests based on arguments
    if [ "$run_all" = true ]; then
        print_info "Running all tests..."
        
        if ! run_unit_tests; then
            exit_code=1
        fi
        
        if ! run_integration_tests; then
            exit_code=1
        fi
        
        if ! run_performance_tests; then
            exit_code=1
        fi
        
    else
        if [ "$run_unit" = true ]; then
            if ! run_unit_tests; then
                exit_code=1
            fi
        fi
        
        if [ "$run_integration" = true ]; then
            if ! run_integration_tests; then
                exit_code=1
            fi
        fi
        
        if [ "$run_performance" = true ]; then
            if ! run_performance_tests; then
                exit_code=1
            fi
        fi
    fi
    
    # Generate report if all tests ran
    if [ "$run_all" = true ] && [ $exit_code -eq 0 ]; then
        generate_test_report
    fi
    
    if [ $exit_code -eq 0 ]; then
        print_header "ALL TESTS COMPLETED SUCCESSFULLY"
        print_success "All tests passed!"
    else
        print_header "SOME TESTS FAILED"
        print_error "Some tests failed. Please check the output above."
    fi
    
    exit $exit_code
}

# Run main function
main "$@" 