package awsfeat

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/models"
)

// Client wraps AWS SDK clients
type Client struct {
	Config aws.Config
	Region string
}

// NewClient creates an AWS client from a stored profile
func NewClient(ctx context.Context, db *storage.Database, profileName string) (*Client, error) {
	// Load profile from database
	var prof models.AWSProfile
	if err := db.Get("aws_profiles", profileName, &prof); err != nil {
		return nil, fmt.Errorf("profile not found: %s", profileName)
	}

	// Build AWS config
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(prof.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			prof.AccessKeyID,
			prof.SecretKey,
			prof.SessionToken,
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		Config: cfg,
		Region: prof.Region,
	}, nil
}
