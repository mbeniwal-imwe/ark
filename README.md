# Ark CLI

A comprehensive command-line tool for developers that provides encrypted credential storage, AWS integration, directory locking, and automation features.

## Features

### 🔐 Encrypted Vault

- Store credentials and secrets securely with AES-256-GCM encryption
- Support for JSON, YAML, and plain text formats
- Search and filter capabilities
- Tag-based organization

### ☁️ AWS Integration

- Manage AWS profiles and credentials
- EC2 instance management (start, stop, SSH, metrics)
- S3 bucket operations (list, navigate, upload, download)
- Automatic credential detection from AWS CLI

### 🔒 Directory Locking

- Lock directories with encryption (macOS)
- Optional directory hiding
- Master password or custom password support

### ⚡ Device Automation

- Caffeinate feature to keep device awake
- Cursor movement automation
- Background daemon process management

### 💾 Backup & Restore

- Encrypted backups to S3
- Point-in-time restore capability
- Client-side encryption

### 📊 Logging & Monitoring

- Comprehensive logging with rotation
- Feature-specific log filtering
- Real-time log viewing

## Installation

### Prerequisites

- macOS (Silicon chip recommended)
- Go 1.23+ (for building from source)

### Quick Install

1. Clone the repository:

```bash
git clone https://github.com/mbeniwal-imwe/ark.git
cd ark
```

2. Build and install:

```bash
make install
```

3. Initialize Ark:

```bash
ark init
```

### Manual Installation

1. Build the binary:

```bash
make build
```

2. Install manually:

```bash
sudo cp build/ark /usr/local/bin/
sudo chmod +x /usr/local/bin/ark
```

3. Add to PATH (if not already done by install script):

```bash
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## Usage

### Initialize Ark

```bash
ark init
```

### Vault Commands

```bash
# Store a credential
ark vault set my-api-key "sk-1234567890abcdef"

# Retrieve a credential
ark vault get my-api-key

# List all credentials
ark vault list

# Search credentials
ark vault search aws

# Delete a credential
ark vault delete old-key
```

### AWS Commands

```bash
# Configure AWS
ark aws configure

# List EC2 instances
ark ec2 list

# Start an instance
ark ec2 start my-instance

# List S3 buckets
ark s3 buckets
```

### Directory Locking

```bash
# Lock a directory
ark lock /path/to/directory

# Unlock a directory
ark unlock /path/to/directory

# List locked directories
ark lock list
```

### Caffeinate

```bash
# Start keeping device awake
ark caffeinate start

# Stop caffeinate
ark caffeinate stop

# Check status
ark caffeinate status
```

## Configuration

Ark stores its configuration in `~/.ark/`:

- `config.yaml` - Main configuration
- `data/ark.db` - Encrypted database
- `logs/` - Application logs
- `backup/` - Backup metadata

## Security

- **Encryption**: AES-256-GCM for all sensitive data
- **Key Derivation**: Argon2id with 100,000 iterations
- **Password Hashing**: PBKDF2 with salt
- **Local Storage**: All data encrypted at rest
- **No Cloud Dependencies**: Works entirely offline

## Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/mbeniwal-imwe/ark.git
cd ark

# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Test GitHub Actions workflow locally
./scripts/test-workflow.sh

# Install
make install
```

### Project Structure

```
ark/
├── cmd/                    # CLI commands
├── internal/
│   ├── core/              # Core functionality
│   ├── storage/           # Data storage
│   ├── features/          # Feature implementations
│   └── ui/                # User interface
├── scripts/               # Installation scripts
└── pkg/                   # Public APIs
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:

- Create an issue on GitHub
- Check the documentation
- Review the logs: `ark logs view`

## Roadmap

- [ ] Linux support
- [ ] Windows support
- [ ] Additional cloud providers
- [ ] Plugin system
- [ ] Web UI
- [ ] Mobile app

## Changelog

### v0.1.0

- Initial release
- Encrypted vault functionality
- AWS integration
- Directory locking (macOS)
- Caffeinate feature
- Backup and restore
