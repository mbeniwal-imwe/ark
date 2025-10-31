#!/bin/bash

# Local workflow testing script
# This script mimics the GitHub Actions workflow for local testing

set -e

echo "🚀 Testing GitHub Actions workflow locally..."
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go 1.23 or later.${NC}"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${BLUE}✓ Go version: $GO_VERSION${NC}"

# Get git commit hash
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
echo -e "${BLUE}✓ Git commit: $GIT_COMMIT${NC}"

echo ""
echo -e "${YELLOW}📦 Step 1: Downloading dependencies...${NC}"
go mod download
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Dependencies downloaded${NC}"
else
    echo -e "${RED}❌ Failed to download dependencies${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}🔍 Step 2: Verifying dependencies...${NC}"
go mod verify
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Dependencies verified${NC}"
else
    echo -e "${RED}❌ Dependency verification failed${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}🧪 Step 3: Running unit tests...${NC}"
go test -v -race -coverprofile=coverage.out ./cmd/...
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}❌ Unit tests failed${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}🧪 Step 4: Running all tests...${NC}"
go test -v ./...
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed${NC}"
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}📊 Step 5: Generating coverage report...${NC}"
go tool cover -func=coverage.out | tail -1
go tool cover -html=coverage.out -o coverage.html 2>/dev/null || true
if [ -f coverage.html ]; then
    echo -e "${GREEN}✓ Coverage report generated: coverage.html${NC}"
fi

echo ""
echo -e "${YELLOW}🔨 Step 6: Building binary...${NC}"
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
mkdir -p build
go build -ldflags "-X github.com/mbeniwal-imwe/ark/cmd.Version=dev -X github.com/mbeniwal-imwe/ark/cmd.BuildDate=$BUILD_DATE -X github.com/mbeniwal-imwe/ark/cmd.GitCommit=$GIT_COMMIT" -o build/ark .
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Binary built successfully${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}✅ Step 7: Testing binary...${NC}"
./build/ark version
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Binary test passed${NC}"
else
    echo -e "${RED}❌ Binary test failed${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}🧹 Step 8: Running lint checks...${NC}"

# Check gofmt
echo "  Checking code formatting..."
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
    echo -e "${RED}❌ The following files are not properly formatted:${NC}"
    echo "$unformatted"
    echo -e "${YELLOW}Run: gofmt -w .${NC}"
    exit 1
else
    echo -e "${GREEN}  ✓ Code formatting OK${NC}"
fi

# Check go vet
echo "  Running go vet..."
go vet ./...
if [ $? -eq 0 ]; then
    echo -e "${GREEN}  ✓ go vet passed${NC}"
else
    echo -e "${RED}  ❌ go vet failed${NC}"
    exit 1
fi

# Check go mod tidy
echo "  Checking go.mod and go.sum..."
go mod tidy
if ! git diff --quiet go.mod go.sum 2>/dev/null; then
    echo -e "${RED}❌ go.mod or go.sum needs to be updated${NC}"
    echo -e "${YELLOW}Run: go mod tidy${NC}"
    git diff go.mod go.sum
    exit 1
else
    echo -e "${GREEN}  ✓ go.mod and go.sum are up to date${NC}"
fi

echo ""
echo -e "${GREEN}🎉 All workflow steps passed!${NC}"
echo ""
echo "Binary location: ./build/ark"
echo "Coverage report: ./coverage.html"
echo ""
echo "You're ready to push! 🚀"

