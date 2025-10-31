package awsfeat

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/models"
)

// EC2Service handles EC2 operations
type EC2Service struct {
	Client *Client
	EC2    *ec2.Client
	DB     *storage.Database
}

// NewEC2Service creates a new EC2 service
func NewEC2Service(ctx context.Context, db *storage.Database, profileName string) (*EC2Service, error) {
	client, err := NewClient(ctx, db, profileName)
	if err != nil {
		return nil, err
	}

	return &EC2Service{
		Client: client,
		EC2:    ec2.NewFromConfig(client.Config),
		DB:     db,
	}, nil
}

// ListInstances lists all EC2 instances
func (s *EC2Service) ListInstances(ctx context.Context) ([]types.Instance, error) {
	result, err := s.EC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	var instances []types.Instance
	for _, reservation := range result.Reservations {
		instances = append(instances, reservation.Instances...)
	}

	return instances, nil
}

// GetInstance retrieves a specific instance by ID
func (s *EC2Service) GetInstance(ctx context.Context, instanceID string) (*types.Instance, error) {
	result, err := s.EC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}

	return &result.Reservations[0].Instances[0], nil
}

// StartInstance starts an EC2 instance
func (s *EC2Service) StartInstance(ctx context.Context, instanceID string) error {
	_, err := s.EC2.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}
	return nil
}

// StopInstance stops an EC2 instance
func (s *EC2Service) StopInstance(ctx context.Context, instanceID string) error {
	_, err := s.EC2.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}
	return nil
}

// RegisterInstance registers an EC2 instance with a custom name in Ark
func (s *EC2Service) RegisterInstance(ctx context.Context, name, instanceID, sshKeyPath, sshUser string) error {
	instance, err := s.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}

	// Create EC2 instance record
	rec := models.NewEC2Instance(name, instanceID, string(instance.InstanceType))
	rec.SetState(string(instance.State.Name))

	// Set IP addresses
	var publicIP, privateIP string
	if instance.PublicIpAddress != nil {
		publicIP = *instance.PublicIpAddress
	}
	if instance.PrivateIpAddress != nil {
		privateIP = *instance.PrivateIpAddress
	}
	rec.SetIPs(publicIP, privateIP)

	// Set SSH config
	if sshKeyPath != "" {
		user := sshUser
		if user == "" {
			user = "ec2-user" // default
		}
		rec.SetSSHConfig(sshKeyPath, user)
	}

	// Store in database
	return s.DB.Set("ec2_instances", name, rec)
}

// GetRegisteredInstance retrieves a registered instance by name
func (s *EC2Service) GetRegisteredInstance(name string) (*models.EC2Instance, error) {
	var rec models.EC2Instance
	if err := s.DB.Get("ec2_instances", name, &rec); err != nil {
		return nil, fmt.Errorf("registered instance not found: %s", name)
	}
	return &rec, nil
}

// ListRegisteredInstances lists all registered instances
func (s *EC2Service) ListRegisteredInstances() ([]models.EC2Instance, error) {
	keys, err := s.DB.List("ec2_instances")
	if err != nil {
		return nil, err
	}

	var instances []models.EC2Instance
	for _, key := range keys {
		var rec models.EC2Instance
		if err := s.DB.Get("ec2_instances", key, &rec); err == nil {
			instances = append(instances, rec)
		}
	}

	return instances, nil
}

// GetInstanceMetrics retrieves CloudWatch metrics for an instance (placeholder)
func (s *EC2Service) GetInstanceMetrics(ctx context.Context, instanceID string) (string, error) {
	// This would require CloudWatch SDK - for now return basic info
	instance, err := s.GetInstance(ctx, instanceID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Instance: %s\nState: %s\nType: %s\nPublic IP: %s\nPrivate IP: %s",
		instanceID,
		instance.State.Name,
		instance.InstanceType,
		getString(instance.PublicIpAddress),
		getString(instance.PrivateIpAddress)), nil
}

func getString(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}

// BuildSSHCommand builds an SSH command for the instance
func BuildSSHCommand(rec *models.EC2Instance) string {
	if rec.SSHKeyPath == "" || rec.PublicIP == "" {
		return ""
	}

	user := rec.SSHUser
	if user == "" {
		user = "ec2-user"
	}

	return fmt.Sprintf("ssh -i %s %s@%s", rec.SSHKeyPath, user, rec.PublicIP)
}
