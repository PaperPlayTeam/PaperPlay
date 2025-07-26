#!/bin/bash

# Configuration
BASE_URL="http://localhost:8080"
TEST_EMAIL="test_$(date +%s)@example.com"
TEST_PASSWORD="password_$(openssl rand -hex 8)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Testing Paper Level API${NC}"
echo "Test User: $TEST_EMAIL"
echo "Test Password: $TEST_PASSWORD"
echo "Base URL: $BASE_URL"
echo

# Step 1: Register user
echo -e "${YELLOW}Step 1: Registering user...${NC}"
register_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\",\"display_name\":\"Test User\"}")

echo "Register Response: $register_response"
echo

# Step 2: Login
echo -e "${YELLOW}Step 2: Logging in...${NC}"
login_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")

access_token=$(echo "$login_response" | jq -r '.access_token')
if [ -z "$access_token" ] || [ "$access_token" = "null" ]; then
    echo -e "${RED}Failed to get access token${NC}"
    echo "Login Response: $login_response"
    exit 1
fi

echo "Login successful, got access token"
echo

# Step 3: Test paper level endpoint
echo -e "${YELLOW}Step 3: Testing GET /api/v1/papers/36283817-9d54-4338-a152-894798c1afc9/level...${NC}"
level_response=$(curl -s -X GET "$BASE_URL/api/v1/papers/36283817-9d54-4338-a152-894798c1afc9/level" \
  -H "Authorization: Bearer $access_token")

echo "Level Response: $level_response"
echo

# Check response status
if echo "$level_response" | jq -e '.error == "paper_not_found"' > /dev/null; then
    echo -e "${YELLOW}Paper not found, checking database...${NC}"
    
    # Use sqlite3 to check papers table
    echo -e "${BLUE}Checking papers table:${NC}"
    sqlite3 data/paperplay.db "SELECT id, subject_id, title FROM papers WHERE id = '36283817-9d54-4338-a152-894798c1afc9';"
    
    echo -e "${BLUE}Listing available papers:${NC}"
    sqlite3 data/paperplay.db "SELECT id, subject_id, title FROM papers LIMIT 5;"
elif echo "$level_response" | jq -e '.error == "level_not_found"' > /dev/null; then
    echo -e "${YELLOW}Level not found, checking database...${NC}"
    
    # Check if paper exists but level doesn't
    echo -e "${BLUE}Checking paper:${NC}"
    sqlite3 data/paperplay.db "SELECT id, subject_id, title FROM papers WHERE id = '36283817-9d54-4338-a152-894798c1afc9';"
    
    echo -e "${BLUE}Checking levels table:${NC}"
    sqlite3 data/paperplay.db "SELECT id, paper_id, name FROM levels WHERE paper_id = '36283817-9d54-4338-a152-894798c1afc9';"
elif echo "$level_response" | jq -e '.success == true' > /dev/null; then
    echo -e "${GREEN}Successfully retrieved level${NC}"
    echo "Level data:"
    echo "$level_response" | jq '.data'
else
    echo -e "${RED}Unexpected response${NC}"
fi

