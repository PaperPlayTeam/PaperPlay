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

    # Handle headers properly by using array to avoid quote issues
    local -a curl_args=(-s -w "\n%{http_code}" -X "$method" "$url")

    if [ -n "$data" ]; then
        curl_args+=(-H "Content-Type: application/json")
    fi

    if [ -n "$headers" ]; then
        # Parse the header string and add to curl_args
        if [[ "$headers" == *"Authorization: Bearer"* ]]; then
            local token=$(echo "$headers" | sed 's/.*Authorization: Bearer \([^"]*\).*/\1/')
            curl_args+=(-H "Authorization: Bearer $token")
        fi
    fi

    if [ -n "$data" ]; then
        curl_args+=(-d "$data")
    fi

    response=$(curl "${curl_args[@]}")

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

    print_test "POST /api/v1/auth/register"
    register_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "$register_data")

    register_http_code=$(echo "$register_response" | tail -n1)
    register_body=$(echo "$register_response" | head -n -1)

    if [ "$register_http_code" = "201" ]; then
        print_success "POST /api/v1/auth/register (Status: $register_http_code)"

        # Extract user info from JSON response
        USER_ID=$(echo "$register_body" | jq -r '.user.id // empty')
        ACCESS_TOKEN=$(echo "$register_body" | jq -r '.access_token // empty')
        REFRESH_TOKEN=$(echo "$register_body" | jq -r '.refresh_token // empty')

        if [ -z "$ACCESS_TOKEN" ]; then
            print_error "Failed to get access token from registration"
            exit 1
        fi
    else
        print_error "POST /api/v1/auth/register (Expected: 201, Got: $register_http_code)"
        echo "Response: $register_body"
        exit 1
    fi

    # User Login
    login_data="{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\"
    }"

    test_http_response "$BASE_URL/api/v1/auth/login" "POST" "$login_data" "" "200" "POST /api/v1/auth/login"

    # Refresh Token (only if we have a refresh token)
    if [ -n "$REFRESH_TOKEN" ]; then
        refresh_data="{
            \"refresh_token\": \"$REFRESH_TOKEN\"
        }"

        test_http_response "$BASE_URL/api/v1/auth/refresh" "POST" "$refresh_data" "" "200" "POST /api/v1/auth/refresh"
    fi

    echo ""
}

# Test user profile endpoints
test_user_profile() {
    print_header "TESTING USER PROFILE ENDPOINTS"

    # Get Initial User Profile
    echo -e "\n${BLUE}Initial User Profile:${NC}"
    print_test "GET /api/v1/users/profile (Initial)"
    profile_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/users/profile" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    profile_http_code=$(echo "$profile_response" | tail -n1)
    profile_body=$(echo "$profile_response" | head -n -1)

    if [ "$profile_http_code" = "200" ]; then
        print_success "GET /api/v1/users/profile (Initial) (Status: $profile_http_code)"
        echo -e "Response: $profile_body"
    else
        print_error "GET /api/v1/users/profile (Initial) (Expected: 200, Got: $profile_http_code)"
        echo -e "Response: $profile_body"
    fi

    # Update User Profile
    echo -e "\n${BLUE}Updating User Profile:${NC}"
    update_data="{
        \"display_name\": \"Updated Test User\",
        \"avatar_url\": \"https://example.com/avatar.jpg\"
    }"

    print_test "PUT /api/v1/users/profile"
    update_response=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/api/v1/users/profile" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "$update_data")

    update_http_code=$(echo "$update_response" | tail -n1)
    update_body=$(echo "$update_response" | head -n -1)

    if [ "$update_http_code" = "200" ]; then
        print_success "PUT /api/v1/users/profile (Status: $update_http_code)"
        echo -e "Response: $update_body"
    else
        print_error "PUT /api/v1/users/profile (Expected: 200, Got: $update_http_code)"
        echo -e "Response: $update_body"
    fi

    # Verify Updated Profile
    echo -e "\n${BLUE}Verifying Updated Profile:${NC}"
    print_test "GET /api/v1/users/profile (Verify)"
    verify_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/users/profile" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    verify_http_code=$(echo "$verify_response" | tail -n1)
    verify_body=$(echo "$verify_response" | head -n -1)

        if [ "$verify_http_code" = "200" ]; then
        print_success "GET /api/v1/users/profile (Verify) (Status: $verify_http_code)"
        echo -e "Response: $verify_body"
        
        # Extract and verify updated values
        updated_name=$(echo "$verify_body" | jq -r '.user.display_name // empty')
        updated_avatar=$(echo "$verify_body" | jq -r '.user.avatar_url // empty')
        
        if [ "$updated_name" = "Updated Test User" ] && [ "$updated_avatar" = "https://example.com/avatar.jpg" ]; then
            echo -e "${GREEN}✓ Profile update verification: Values match${NC}"
            echo -e "  - display_name: $updated_name"
            echo -e "  - avatar_url: $updated_avatar"
        else
            echo -e "${RED}✗ Profile update verification: Values don't match${NC}"
            echo -e "  - Expected display_name: Updated Test User"
            echo -e "  - Got display_name: $updated_name"
            echo -e "  - Expected avatar_url: https://example.com/avatar.jpg"
            echo -e "  - Got avatar_url: $updated_avatar"
        fi
    else
        print_error "GET /api/v1/users/profile (Verify) (Expected: 200, Got: $verify_http_code)"
        echo -e "Response: $verify_body"
    fi

    # Get User Progress
    echo -e "\n${BLUE}User Progress:${NC}"
    print_test "GET /api/v1/users/progress"
    progress_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/users/progress" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    progress_http_code=$(echo "$progress_response" | tail -n1)
    progress_body=$(echo "$progress_response" | head -n -1)

    if [ "$progress_http_code" = "200" ]; then
        print_success "GET /api/v1/users/progress (Status: $progress_http_code)"
        echo -e "Response: $progress_body"
    else
        print_error "GET /api/v1/users/progress (Expected: 200, Got: $progress_http_code)"
        echo -e "Response: $progress_body"
    fi

    # Get User Achievements
    echo -e "\n${BLUE}User Achievements:${NC}"
    print_test "GET /api/v1/users/achievements"
    achievements_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/users/achievements" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    achievements_http_code=$(echo "$achievements_response" | tail -n1)
    achievements_body=$(echo "$achievements_response" | head -n -1)

    if [ "$achievements_http_code" = "200" ]; then
        print_success "GET /api/v1/users/achievements (Status: $achievements_http_code)"
        echo -e "Response: $achievements_body"
    else
        print_error "GET /api/v1/users/achievements (Expected: 200, Got: $achievements_http_code)"
        echo -e "Response: $achievements_body"
    fi

    echo ""
}

# Test level system endpoints
test_level_system() {
    print_header "TESTING LEVEL SYSTEM ENDPOINTS"

    auth_header="-H \"Authorization: Bearer $ACCESS_TOKEN\""

    # Get All Subjects
    print_test "GET /api/v1/subjects"
    subjects_full_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/subjects" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    subjects_http_code=$(echo "$subjects_full_response" | tail -n1)
    subjects_body=$(echo "$subjects_full_response" | head -n -1)

    if [ "$subjects_http_code" = "200" ]; then
        print_success "GET /api/v1/subjects (Status: $subjects_http_code)"
        echo -e "Response: $subjects_body"
        # Extract first subject ID if exists
        SUBJECT_ID=$(echo "$subjects_body" | jq -r '.data[0].id // empty')
    else
        print_error "GET /api/v1/subjects (Expected: 200, Got: $subjects_http_code)"
        echo -e "Response: $subjects_body"
    fi

    if [ -n "$SUBJECT_ID" ]; then
        # Get Single Subject
        test_http_response "$BASE_URL/api/v1/subjects/$SUBJECT_ID" "GET" "" "$auth_header" "200" "GET /api/v1/subjects/{subject_id}"

        # Get Subject Papers
        print_test "GET /api/v1/subjects/{subject_id}/papers"
        papers_full_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/subjects/$SUBJECT_ID/papers" \
            -H "Authorization: Bearer $ACCESS_TOKEN")
        papers_http_code=$(echo "$papers_full_response" | tail -n1)
        papers_body=$(echo "$papers_full_response" | head -n -1)

        if [ "$papers_http_code" = "200" ]; then
            print_success "GET /api/v1/subjects/{subject_id}/papers (Status: $papers_http_code)"
            echo -e "Response: $papers_body"
            # Extract first paper ID if exists
            PAPER_ID=$(echo "$papers_body" | jq -r '.data[0].id // empty')
        else
            print_error "GET /api/v1/subjects/{subject_id}/papers (Expected: 200, Got: $papers_http_code)"
            echo -e "Response: $papers_body"
        fi

        # Get Subject Roadmap
        test_http_response "$BASE_URL/api/v1/subjects/$SUBJECT_ID/roadmap" "GET" "" "$auth_header" "200" "GET /api/v1/subjects/{subject_id}/roadmap"

        if [ -n "$PAPER_ID" ]; then
            # Get Single Paper
            test_http_response "$BASE_URL/api/v1/papers/$PAPER_ID" "GET" "" "$auth_header" "200" "GET /api/v1/papers/{paper_id}"

            # Get Paper Level
            print_test "GET /api/v1/papers/{paper_id}/level"
            level_full_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/papers/$PAPER_ID/level" \
                -H "Authorization: Bearer $ACCESS_TOKEN")
            level_http_code=$(echo "$level_full_response" | tail -n1)
            level_body=$(echo "$level_full_response" | head -n -1)

            if [ "$level_http_code" = "200" ]; then
                print_success "GET /api/v1/papers/{paper_id}/level (Status: $level_http_code)"
                # Extract level ID if exists
                LEVEL_ID=$(echo "$level_body" | jq -r '.data.id // empty')
            else
                print_error "GET /api/v1/papers/{paper_id}/level (Expected: 200, Got: $level_http_code)"
                echo "Response: $level_body"
            fi

            if [ -n "$LEVEL_ID" ]; then
                # Get Single Level
                test_http_response "$BASE_URL/api/v1/levels/$LEVEL_ID" "GET" "" "$auth_header" "200" "GET /api/v1/levels/{level_id}"

                # Get Level Questions
                print_test "GET /api/v1/levels/{level_id}/questions"
                questions_full_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/levels/$LEVEL_ID/questions" \
                    -H "Authorization: Bearer $ACCESS_TOKEN")
                questions_http_code=$(echo "$questions_full_response" | tail -n1)
                questions_body=$(echo "$questions_full_response" | head -n -1)

                if [ "$questions_http_code" = "200" ]; then
                    print_success "GET /api/v1/levels/{level_id}/questions (Status: $questions_http_code)"
                    # Extract first question ID if exists
                    QUESTION_ID=$(echo "$questions_body" | jq -r '.data[0].id // empty')
                else
                    print_error "GET /api/v1/levels/{level_id}/questions (Expected: 200, Got: $questions_http_code)"
                    echo "Response: $questions_body"
                fi

                # Start Level
                test_http_response "$BASE_URL/api/v1/levels/$LEVEL_ID/start" "POST" "" "$auth_header" "200" "POST /api/v1/levels/{level_id}/start"

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
    curl -s "$BASE_URL/health" >/dev/null
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
        curl -s "$BASE_URL/health" >/dev/null &
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
    if ! command -v jq &>/dev/null; then
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
