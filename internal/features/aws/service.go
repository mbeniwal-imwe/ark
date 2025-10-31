package awsfeat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/models"
)

type Service struct {
	DB *storage.Database
}

// ImportFromAWSDir parses ~/.aws/credentials and ~/.aws/config and stores profiles
func (s *Service) ImportFromAWSDir(home string) (int, error) {
	awsDir := filepath.Join(home, ".aws")
	credsPath := filepath.Join(awsDir, "credentials")
	configPath := filepath.Join(awsDir, "config")

	profiles := map[string]*models.AWSProfile{}

	// credentials file
	if data, err := os.ReadFile(credsPath); err == nil {
		parseIni(string(data), func(section string, kv map[string]string) {
			name := section
			p := profiles[name]
			if p == nil {
				p = models.NewAWSProfile(name, "", "", "")
				profiles[name] = p
			}
			if v := kv["aws_access_key_id"]; v != "" {
				p.AccessKeyID = v
			}
			if v := kv["aws_secret_access_key"]; v != "" {
				p.SecretKey = v
			}
			if v := kv["aws_session_token"]; v != "" {
				p.SessionToken = v
			}
		})
	}

	// config file (region, output)
	if data, err := os.ReadFile(configPath); err == nil {
		parseIni(string(data), func(section string, kv map[string]string) {
			name := strings.TrimPrefix(section, "profile ")
			p := profiles[name]
			if p == nil {
				p = models.NewAWSProfile(name, "", "", "")
				profiles[name] = p
			}
			if v := kv["region"]; v != "" {
				p.Region = v
			}
			if v := kv["output"]; v != "" {
				p.Output = v
			}
		})
	}

	// Persist to DB
	count := 0
	for name, prof := range profiles {
		if prof.AccessKeyID == "" || prof.SecretKey == "" {
			continue
		}
		if err := s.DB.Set("aws_profiles", name, prof); err == nil {
			count++
		}
	}
	return count, nil
}

func (s *Service) ListProfiles() ([]models.AWSProfile, error) {
	keys, err := s.DB.List("aws_profiles")
	if err != nil {
		return nil, err
	}
	var out []models.AWSProfile
	for _, k := range keys {
		var p models.AWSProfile
		if err := s.DB.Get("aws_profiles", k, &p); err == nil {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *Service) SetDefaultProfile(name string) error {
	// Store default profile name in config bucket
	return s.DB.Set("config", "aws_default_profile", map[string]string{"name": name})
}

func (s *Service) GetDefaultProfile() (string, error) {
	var v map[string]string
	if err := s.DB.Get("config", "aws_default_profile", &v); err != nil {
		return "", err
	}
	return v["name"], nil
}

// TestConnection attempts to validate credentials using AWS STS
func (s *Service) TestConnection(ctx context.Context, profile string) (string, error) {
	client, err := NewClient(ctx, s.DB, profile)
	if err != nil {
		return "", err
	}

	// Use STS GetCallerIdentity to test connection
	stsClient := sts.NewFromConfig(client.Config)
	result, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("connection test failed: %w", err)
	}

	return fmt.Sprintf("✅ Connection successful\nAccount: %s\nUser ARN: %s\nRegion: %s",
		aws.ToString(result.Account),
		aws.ToString(result.Arn),
		client.Region), nil
}

// --- simple INI parser (minimal) ---
func parseIni(content string, onSection func(section string, kv map[string]string)) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	current := ""
	kv := map[string]string{}
	flush := func() {
		if current != "" {
			onSection(current, kv)
		}
		kv = map[string]string{}
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			flush()
			current = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		if eq := strings.IndexByte(line, '='); eq != -1 {
			k := strings.TrimSpace(line[:eq])
			v := strings.TrimSpace(line[eq+1:])
			kv[strings.ToLower(k)] = strings.Trim(v, "\"")
		}
	}
	flush()
}
