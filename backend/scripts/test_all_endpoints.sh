#!/bin/bash

# PaperPlay API Integration Test Suite
# Tests all API endpoints for functionality and performance

set -e

# Configuration
BASE_URL="http://localhost:8080"
TEST_EMAIL="test_$(date +%s)@example.com"
TEST_PASSWORD="password123"
TEST_DISPLAY_NAME="Test User $(date +%s)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Global variables
ACCESS_TOKEN=""
REFRESH_TOKEN=""
USER_ID=""

# Utility functions
print_header() {
    echo -e "${BLUE}=================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=================================${NC}"
}

print_test() {
    echo -e "${YELLOW}Testing: $1${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

print_success() {
    echo -e "${GREEN}✓ PASS: $1${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

print_error() {
    echo -e "${RED}✗ FAIL: $1${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

print_summary() {
    echo -e "${BLUE}=================================${NC}"
    echo -e "${BLUE}TEST SUMMARY${NC}"
    echo -e "${BLUE}=================================${NC}"
    echo -e "Total Tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed!${NC}"
        exit 1
    fi
}

# Helper function to test HTTP response
test_http_response() {
    local url="$1"
    local method="$2"
    local data="$3"
    local headers="$4"
    local expected_status="$5"
    local test_name="$6"
    
    print_test "$test_name"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            $headers \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            $headers)
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_status" ]; then
        print_success "$test_name (Status: $http_code)"
        echo "$body"
        return 0
    else
        print_error "$test_name (Expected: $expected_status, Got: $http_code)"
        echo "Response: $body"
        return 1
    fi
}

# Check if server is running
check_server() {
    print_header "CHECKING SERVER STATUS"
    
    if ! test_http_response "$BASE_URL/health" "GET" "" "" "200" "Server Health Check"; then
        echo -e "${RED}Server is not running or not healthy. Please start the server first.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Server is running and healthy${NC}"
    echo ""
}

# Test system endpoints
test_system_endpoints() {
    print_header "TESTING SYSTEM ENDPOINTS"
    
    # Health check
    test_http_response "$BASE_URL/health" "GET" "" "" "200" "GET /health"
    
    # Metrics endpoint
    test_http_response "$BASE_URL/metrics" "GET" "" "" "200" "GET /metrics"
    
    echo ""
}

# Test user authentication
test_user_authentication() {
    print_header "TESTING USER AUTHENTICATION"
    
    # User Registration
    register_data="{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\",
        \"display_name\": \"$TEST_DISPLAY_NAME\"
    }"
    
    register_response=$(test_http_response "$BASE_URL/api/v1/auth/register" "POST" "$register_data" "" "201" "POST /api/v1/auth/register")
    
    # Extract user info
    USER_ID=$(echo "$register_response" | jq -r '.user.id // empty')
    ACCESS_TOKEN=$(echo "$register_response" | jq -r '.access_token // empty')
    REFRESH_TOKEN=$(echo "$register_response" | jq -r '.refresh_token // empty')
    
    if [ -z "$ACCESS_TOKEN" ]; then
        print_error "Failed to get access token from registration"
        exit 1
    fi
    
    # User Login
    login_data="{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\"
    }"
    
    test_http_response "$BASE_URL/api/v1/auth/login" "POST" "$login_data" "" "200" "POST /api/v1/auth/login"
    
    # Refresh Token
    refresh_data="{
        \"refresh_token\": \"$REFRESH_TOKEN\"
    }"
    
    test_http_response "$BASE_URL/api/v1/auth/refresh" "POST" "$refresh_data" "" "200" "POST /api/v1/auth/refresh"
    
    echo ""
}

# Test user profile endpoints
test_user_profile() {
    print_header "TESTING USER PROFILE ENDPOINTS"
    
    auth_header="-H \"Authorization: Bearer $ACCESS_TOKEN\""
    
    # Get User Profile
    test_http_response "$BASE_URL/api/v1/users/profile" "GET" "" "$auth_header" "200" "GET /api/v1/users/profile"
    
    # Update User Profile
    update_data="{
        \"display_name\": \"Updated Test User\",
        \"avatar_url\": \"https://example.com/avatar.jpg\"
    }"
    
    test_http_response "$BASE_URL/api/v1/users/profile" "PUT" "$update_data" "$auth_header" "200" "PUT /api/v1/users/profile"
    
    # Get User Progress
    test_http_response "$BASE_URL/api/v1/users/progress" "GET" "" "$auth_header" "200" "GET /api/v1/users/progress"
    
    # Get User Achievements
    test_http_response "$BASE_URL/api/v1/users/achievements" "GET" "" "$auth_header" "200" "GET /api/v1/users/achievements"
    
    echo ""
}

# Test level system endpoints
test_level_system() {
    print_header "TESTING LEVEL SYSTEM ENDPOINTS"
    
    auth_header="-H \"Authorization: Bearer $ACCESS_TOKEN\""
    
    # Get All Subjects
    subjects_response=$(test_http_response "$BASE_URL/api/v1/subjects" "GET" "" "$auth_header" "200" "GET /api/v1/subjects")
    
    # Extract first subject ID if exists
    SUBJECT_ID=$(echo "$subjects_response" | jq -r '.data[0].id // empty')
    
    if [ -n "$SUBJECT_ID" ]; then
        # Get Single Subject
        test_http_response "$BASE_URL/api/v1/subjects/$SUBJECT_ID" "GET" "" "$auth_header" "200" "GET /api/v1/subjects/{subject_id}"
        
        # Get Subject Papers
        papers_response=$(test_http_response "$BASE_URL/api/v1/subjects/$SUBJECT_ID/papers" "GET" "" "$auth_header" "200" "GET /api/v1/subjects/{subject_id}/papers")
        
        # Get Subject Roadmap
        test_http_response "$BASE_URL/api/v1/subjects/$SUBJECT_ID/roadmap" "GET" "" "$auth_header" "200" "GET /api/v1/subjects/{subject_id}/roadmap"
        
        # Extract first paper ID if exists
        PAPER_ID=$(echo "$papers_response" | jq -r '.data[0].id // empty')
        
        if [ -n "$PAPER_ID" ]; then
            # Get Single Paper
            test_http_response "$BASE_URL/api/v1/papers/$PAPER_ID" "GET" "" "$auth_header" "200" "GET /api/v1/papers/{paper_id}"
            
            # Get Paper Level
            level_response=$(test_http_response "$BASE_URL/api/v1/papers/$PAPER_ID/level" "GET" "" "$auth_header" "200" "GET /api/v1/papers/{paper_id}/level")
            
            # Extract level ID if exists
            LEVEL_ID=$(echo "$level_response" | jq -r '.data.id // empty')
            
            if [ -n "$LEVEL_ID" ]; then
                # Get Single Level
                test_http_response "$BASE_URL/api/v1/levels/$LEVEL_ID" "GET" "" "$auth_header" "200" "GET /api/v1/levels/{level_id}"
                
                # Get Level Questions
                questions_response=$(test_http_response "$BASE_URL/api/v1/levels/$LEVEL_ID/questions" "GET" "" "$auth_header" "200" "GET /api/v1/levels/{level_id}/questions")
                
                # Start Level
                test_http_response "$BASE_URL/api/v1/levels/$LEVEL_ID/start" "POST" "" "$auth_header" "200" "POST /api/v1/levels/{level_id}/start"
                
                # Extract first question ID if exists
                QUESTION_ID=$(echo "$questions_response" | jq -r '.data[0].id // empty')
                
                if [ -n "$QUESTION_ID" ]; then
                    # Get Single Question
                    test_http_response "$BASE_URL/api/v1/questions/$QUESTION_ID" "GET" "" "$auth_header" "200" "GET /api/v1/questions/{question_id}"
                    
                    # Submit Answer
                    submit_data="{
                        \"question_id\": \"$QUESTION_ID\",
                        \"answer_json\": \"Option A\",
                        \"duration_ms\": 30000
                    }"
                    
                    test_http_response "$BASE_URL/api/v1/levels/$LEVEL_ID/submit" "POST" "$submit_data" "$auth_header" "200" "POST /api/v1/levels/{level_id}/submit"
                fi
                
                # Complete Level
                test_http_response "$BASE_URL/api/v1/levels/$LEVEL_ID/complete" "POST" "" "$auth_header" "200" "POST /api/v1/levels/{level_id}/complete"
            fi
        fi
    else
        print_error "No subjects found in database. Skipping level system tests."
    fi
    
    echo ""
}

# Test achievement system endpoints
test_achievement_system() {
    print_header "TESTING ACHIEVEMENT SYSTEM ENDPOINTS"
    
    auth_header="-H \"Authorization: Bearer $ACCESS_TOKEN\""
    
    # Get All Achievements
    test_http_response "$BASE_URL/api/v1/achievements" "GET" "" "$auth_header" "200" "GET /api/v1/achievements"
    
    # Get User Achievements
    test_http_response "$BASE_URL/api/v1/achievements/user" "GET" "" "$auth_header" "200" "GET /api/v1/achievements/user"
    
    # Trigger Achievement Evaluation
    test_http_response "$BASE_URL/api/v1/achievements/evaluate" "POST" "" "$auth_header" "200" "POST /api/v1/achievements/evaluate"
    
    echo ""
}

# Test WebSocket and system stats
test_websocket_and_stats() {
    print_header "TESTING WEBSOCKET AND SYSTEM STATS"
    
    auth_header="-H \"Authorization: Bearer $ACCESS_TOKEN\""
    
    # Get WebSocket Stats
    test_http_response "$BASE_URL/api/v1/ws/stats" "GET" "" "$auth_header" "200" "GET /api/v1/ws/stats"
    
    # Get System Stats
    test_http_response "$BASE_URL/api/v1/system/stats" "GET" "" "$auth_header" "200" "GET /api/v1/system/stats"
    
    echo ""
}

# Test error cases
test_error_cases() {
    print_header "TESTING ERROR CASES"
    
    # Test unauthorized access
    test_http_response "$BASE_URL/api/v1/users/profile" "GET" "" "" "401" "Unauthorized access to protected endpoint"
    
    # Test invalid login
    invalid_login_data="{
        \"email\": \"nonexistent@example.com\",
        \"password\": \"wrongpassword\"
    }"
    
    test_http_response "$BASE_URL/api/v1/auth/login" "POST" "$invalid_login_data" "" "401" "Invalid login credentials"
    
    # Test invalid registration
    invalid_register_data="{
        \"email\": \"invalid-email\",
        \"password\": \"123\",
        \"display_name\": \"\"
    }"
    
    test_http_response "$BASE_URL/api/v1/auth/register" "POST" "$invalid_register_data" "" "400" "Invalid registration data"
    
    # Test non-existent resource
    auth_header="-H \"Authorization: Bearer $ACCESS_TOKEN\""
    test_http_response "$BASE_URL/api/v1/subjects/nonexistent-id" "GET" "" "$auth_header" "404" "Non-existent subject"
    
    echo ""
}

# Test user logout
test_user_logout() {
    print_header "TESTING USER LOGOUT"
    
    auth_header="-H \"Authorization: Bearer $ACCESS_TOKEN\""
    
    # User Logout
    test_http_response "$BASE_URL/api/v1/users/logout" "POST" "" "$auth_header" "200" "POST /api/v1/users/logout"
    
    echo ""
}

# Performance tests
test_performance() {
    print_header "TESTING PERFORMANCE"
    
    # Test health endpoint response time
    print_test "Health endpoint response time"
    start_time=$(date +%s%N)
    curl -s "$BASE_URL/health" > /dev/null
    end_time=$(date +%s%N)
    response_time=$(((end_time - start_time) / 1000000))
    
    if [ $response_time -lt 100 ]; then
        print_success "Health endpoint response time: ${response_time}ms (< 100ms)"
    elif [ $response_time -lt 500 ]; then
        print_success "Health endpoint response time: ${response_time}ms (< 500ms)"
    else
        print_error "Health endpoint response time: ${response_time}ms (>= 500ms)"
    fi
    
    # Test concurrent requests
    print_test "Concurrent health requests"
    concurrent_start=$(date +%s%N)
    for i in {1..5}; do
        curl -s "$BASE_URL/health" > /dev/null &
    done
    wait
    concurrent_end=$(date +%s%N)
    concurrent_time=$(((concurrent_end - concurrent_start) / 1000000))
    
    if [ $concurrent_time -lt 1000 ]; then
        print_success "5 concurrent requests completed in ${concurrent_time}ms"
    else
        print_error "5 concurrent requests took ${concurrent_time}ms (too slow)"
    fi
    
    echo ""
}

# Main test execution
main() {
    echo -e "${BLUE}Starting PaperPlay API Integration Tests${NC}"
    echo -e "${BLUE}Test User: $TEST_EMAIL${NC}"
    echo -e "${BLUE}Base URL: $BASE_URL${NC}"
    echo ""
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}jq is required but not installed. Please install jq to run these tests.${NC}"
        exit 1
    fi
    
    # Run all test suites
    check_server
    test_system_endpoints
    test_user_authentication
    test_user_profile
    test_level_system
    test_achievement_system
    test_websocket_and_stats
    test_error_cases
    test_user_logout
    test_performance
    
    # Print summary
    print_summary
}

# Run main function
main "$@" 