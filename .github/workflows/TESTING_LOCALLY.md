# Testing GitHub Actions Workflow Locally

There are two ways to test the GitHub Actions workflow locally before pushing:

## Method 1: Using the Test Script (Recommended)

The easiest way is to use the provided test script that mimics the GitHub Actions workflow:

```bash
./scripts/test-workflow.sh
```

This script will:

- ✅ Check Go version
- ✅ Download and verify dependencies
- ✅ Run all unit tests
- ✅ Generate coverage reports
- ✅ Build the binary
- ✅ Test the binary
- ✅ Run lint checks (gofmt, go vet, go mod tidy)

If all steps pass, you're ready to push!

## Method 2: Using `act` (Full GitHub Actions Simulation)

You can use [act](https://github.com/nektos/act) to run GitHub Actions workflows locally:

### Installation

**macOS:**

```bash
brew install act
```

**Linux:**

```bash
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

**Windows (via Docker):**
Download from [act releases](https://github.com/nektos/act/releases)

### Usage

```bash
# Run the CI workflow
act pull_request

# Or simulate a push event
act push

# Run specific jobs
act -j test
act -j lint
act -j build

# Run with verbose output
act -v pull_request
```

**Note:** `act` requires Docker to be running. It may not perfectly replicate GitHub Actions, but it's good for catching most issues.

## Method 3: Manual Testing

You can also manually run the commands from the workflow:

```bash
# Download dependencies
go mod download
go mod verify

# Run tests
go test -v -race -coverprofile=coverage.out ./cmd/...
go test -v ./...

# Check formatting
gofmt -l .
go vet ./...

# Check go.mod
go mod tidy
git diff go.mod go.sum  # Should be empty

# Build
mkdir -p build
GIT_COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
go build -ldflags "-X github.com/mbeniwal-imwe/ark/cmd.Version=dev -X github.com/mbeniwal-imwe/ark/cmd.BuildDate=$BUILD_DATE -X github.com/mbeniwal-imwe/ark/cmd.GitCommit=$GIT_COMMIT" -o build/ark .

# Test binary
./build/ark version
```

## Quick Pre-Push Checklist

Before pushing, make sure:

- [ ] All tests pass: `make test` or `go test ./...`
- [ ] Code is formatted: `gofmt -w .`
- [ ] No lint errors: `go vet ./...`
- [ ] go.mod is tidy: `go mod tidy`
- [ ] Binary builds: `make build`
- [ ] Binary works: `./build/ark version`

Run `./scripts/test-workflow.sh` to check all of these automatically!
