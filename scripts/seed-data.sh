#!/bin/bash
# Seed Development Data Script
# Populates the database with test data for development

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
POSTGRES_DSN="${POSTGRES_DSN:-postgres://library:library123@localhost:5432/library?sslmode=disable}"

echo -e "${YELLOW}🌱 Seeding development data...${NC}"
echo ""

#######################################
# Helper Functions
#######################################

api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4

    if [ -n "$token" ]; then
        curl -s -X "$method" "${API_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d "$data"
    else
        curl -s -X "$method" "${API_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$data"
    fi
}

extract_json_field() {
    local json=$1
    local field=$2
    echo "$json" | grep -o "\"$field\":\"[^\"]*\"" | cut -d'"' -f4
}

#######################################
# Check if API is running
#######################################

echo -e "${YELLOW}Checking if API server is running...${NC}"
if ! curl -s "${API_URL}/health" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  API server is not running at ${API_URL}${NC}"
    echo "   Starting API server in background..."

    # Build if needed
    if [ ! -f "bin/library-api" ]; then
        make build-api > /dev/null 2>&1
    fi

    # Start server in background
    POSTGRES_DSN="$POSTGRES_DSN" bin/library-api &
    API_PID=$!

    # Wait for server to be ready
    max_attempts=30
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if curl -s "${API_URL}/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ API server started${NC}"
            break
        fi
        attempt=$((attempt + 1))
        sleep 1
    done

    if [ $attempt -eq $max_attempts ]; then
        echo -e "${RED}❌ Failed to start API server${NC}"
        kill $API_PID 2>/dev/null || true
        exit 1
    fi

    # Remember to stop server at the end
    STOP_API=true
else
    echo -e "${GREEN}✓ API server is running${NC}"
    STOP_API=false
fi
echo ""

#######################################
# Create Test Users
#######################################

echo -e "${YELLOW}👥 Creating test users...${NC}"

# Admin user
echo "  Creating admin user..."
ADMIN_RESPONSE=$(api_call POST "/api/v1/auth/register" '{
    "email": "admin@library.com",
    "password": "Admin123!@#",
    "full_name": "Admin User"
}')
ADMIN_TOKEN=$(extract_json_field "$ADMIN_RESPONSE" "access_token")

if [ -n "$ADMIN_TOKEN" ]; then
    echo -e "    ${GREEN}✓ admin@library.com (password: Admin123!@#)${NC}"
else
    echo -e "    ${YELLOW}⚠️  admin@library.com (may already exist)${NC}"
fi

# Regular user
echo "  Creating regular user..."
USER_RESPONSE=$(api_call POST "/api/v1/auth/register" '{
    "email": "user@library.com",
    "password": "User123!@#",
    "full_name": "Regular User"
}')
USER_TOKEN=$(extract_json_field "$USER_RESPONSE" "access_token")

if [ -n "$USER_TOKEN" ]; then
    echo -e "    ${GREEN}✓ user@library.com (password: User123!@#)${NC}"
else
    echo -e "    ${YELLOW}⚠️  user@library.com (may already exist)${NC}"
    # Try to login
    USER_RESPONSE=$(api_call POST "/api/v1/auth/login" '{
        "email": "user@library.com",
        "password": "User123!@#"
    }')
    USER_TOKEN=$(extract_json_field "$USER_RESPONSE" "access_token")
fi

# Premium user
echo "  Creating premium user..."
PREMIUM_RESPONSE=$(api_call POST "/api/v1/auth/register" '{
    "email": "premium@library.com",
    "password": "Premium123!@#",
    "full_name": "Premium Member"
}')
PREMIUM_TOKEN=$(extract_json_field "$PREMIUM_RESPONSE" "access_token")

if [ -n "$PREMIUM_TOKEN" ]; then
    echo -e "    ${GREEN}✓ premium@library.com (password: Premium123!@#)${NC}"
else
    echo -e "    ${YELLOW}⚠️  premium@library.com (may already exist)${NC}"
fi

echo ""

#######################################
# Create Sample Books
#######################################

echo -e "${YELLOW}📚 Creating sample books...${NC}"

# Use admin token if available, otherwise user token
TOKEN="${ADMIN_TOKEN:-$USER_TOKEN}"

if [ -z "$TOKEN" ]; then
    echo -e "${RED}❌ No authentication token available${NC}"
    exit 1
fi

# Book 1: Clean Code
echo "  Creating 'Clean Code' by Robert C. Martin..."
api_call POST "/api/v1/books" '{
    "title": "Clean Code: A Handbook of Agile Software Craftsmanship",
    "isbn": "978-0132350884",
    "published_year": 2008,
    "genre": "Programming",
    "quantity": 5,
    "authors": [
        {
            "name": "Robert C. Martin",
            "biography": "Robert Cecil Martin, known colloquially as Uncle Bob, is an American software engineer and author."
        }
    ]
}' "$TOKEN" > /dev/null
echo -e "    ${GREEN}✓ Clean Code${NC}"

# Book 2: Design Patterns
echo "  Creating 'Design Patterns' by Gang of Four..."
api_call POST "/api/v1/books" '{
    "title": "Design Patterns: Elements of Reusable Object-Oriented Software",
    "isbn": "978-0201633610",
    "published_year": 1994,
    "genre": "Software Engineering",
    "quantity": 3,
    "authors": [
        {
            "name": "Erich Gamma",
            "biography": "Swiss computer scientist and co-author of the influential software engineering textbook, Design Patterns."
        },
        {
            "name": "Richard Helm",
            "biography": "Australian computer scientist known as one of the Gang of Four."
        },
        {
            "name": "Ralph Johnson",
            "biography": "American computer scientist and professor at the University of Illinois."
        },
        {
            "name": "John Vlissides",
            "biography": "Software scientist known for his contributions to design patterns."
        }
    ]
}' "$TOKEN" > /dev/null
echo -e "    ${GREEN}✓ Design Patterns${NC}"

# Book 3: The Pragmatic Programmer
echo "  Creating 'The Pragmatic Programmer'..."
api_call POST "/api/v1/books" '{
    "title": "The Pragmatic Programmer: Your Journey To Mastery",
    "isbn": "978-0135957059",
    "published_year": 2019,
    "genre": "Programming",
    "quantity": 4,
    "authors": [
        {
            "name": "David Thomas",
            "biography": "Programmer and author, co-author of The Pragmatic Programmer."
        },
        {
            "name": "Andrew Hunt",
            "biography": "Software engineer and one of the 17 original signatories of the Agile Manifesto."
        }
    ]
}' "$TOKEN" > /dev/null
echo -e "    ${GREEN}✓ The Pragmatic Programmer${NC}"

# Book 4: Refactoring
echo "  Creating 'Refactoring' by Martin Fowler..."
api_call POST "/api/v1/books" '{
    "title": "Refactoring: Improving the Design of Existing Code",
    "isbn": "978-0134757599",
    "published_year": 2018,
    "genre": "Software Engineering",
    "quantity": 3,
    "authors": [
        {
            "name": "Martin Fowler",
            "biography": "British software developer, author and international public speaker on software development."
        }
    ]
}' "$TOKEN" > /dev/null
echo -e "    ${GREEN}✓ Refactoring${NC}"

# Book 5: Domain-Driven Design
echo "  Creating 'Domain-Driven Design'..."
api_call POST "/api/v1/books" '{
    "title": "Domain-Driven Design: Tackling Complexity in the Heart of Software",
    "isbn": "978-0321125217",
    "published_year": 2003,
    "genre": "Software Architecture",
    "quantity": 2,
    "authors": [
        {
            "name": "Eric Evans",
            "biography": "American software engineer and author, known for Domain-Driven Design."
        }
    ]
}' "$TOKEN" > /dev/null
echo -e "    ${GREEN}✓ Domain-Driven Design${NC}"

# Book 6: Clean Architecture
echo "  Creating 'Clean Architecture'..."
api_call POST "/api/v1/books" '{
    "title": "Clean Architecture: A Craftsman Guide to Software Structure and Design",
    "isbn": "978-0134494166",
    "published_year": 2017,
    "genre": "Software Architecture",
    "quantity": 4,
    "authors": [
        {
            "name": "Robert C. Martin",
            "biography": "Robert Cecil Martin, known colloquially as Uncle Bob, is an American software engineer and author."
        }
    ]
}' "$TOKEN" > /dev/null
echo -e "    ${GREEN}✓ Clean Architecture${NC}"

echo ""

#######################################
# Summary
#######################################

echo -e "${GREEN}========================================"
echo " ✅ Seed Data Complete!"
echo "========================================${NC}"
echo ""
echo "Created test accounts:"
echo "  • admin@library.com     (Admin123!@#)"
echo "  • user@library.com      (User123!@#)"
echo "  • premium@library.com   (Premium123!@#)"
echo ""
echo "Created sample books:"
echo "  • Clean Code"
echo "  • Design Patterns"
echo "  • The Pragmatic Programmer"
echo "  • Refactoring"
echo "  • Domain-Driven Design"
echo "  • Clean Architecture"
echo ""
echo "You can now:"
echo "  • Login with any test account"
echo "  • Browse books via API or Swagger UI"
echo "  • Test reservations and other features"
echo ""

# Stop API if we started it
if [ "$STOP_API" = true ]; then
    echo -e "${YELLOW}Stopping API server...${NC}"
    kill $API_PID 2>/dev/null || true
    echo -e "${GREEN}✓ API server stopped${NC}"
fi
