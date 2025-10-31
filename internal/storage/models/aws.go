package models

import (
	"time"
)

// AWSProfile represents an AWS profile configuration
type AWSProfile struct {
	Name         string            `json:"name"`
	AccessKeyID  string            `json:"access_key_id"`
	SecretKey    string            `json:"secret_key"`
	SessionToken string            `json:"session_token,omitempty"`
	Region       string            `json:"region"`
	Output       string            `json:"output"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// NewAWSProfile creates a new AWS profile
func NewAWSProfile(name, accessKeyID, secretKey, region string) *AWSProfile {
	now := time.Now()
	return &AWSProfile{
		Name:        name,
		AccessKeyID: accessKeyID,
		SecretKey:   secretKey,
		Region:      region,
		Output:      "json",
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    make(map[string]string),
	}
}

// SetSessionToken sets the session token for the profile
func (p *AWSProfile) SetSessionToken(token string) {
	p.SessionToken = token
	p.UpdatedAt = time.Now()
}

// SetOutput sets the output format for the profile
func (p *AWSProfile) SetOutput(output string) {
	p.Output = output
	p.UpdatedAt = time.Now()
}

// SetMetadata sets metadata for the profile
func (p *AWSProfile) SetMetadata(key, value string) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	p.Metadata[key] = value
	p.UpdatedAt = time.Now()
}

// EC2Instance represents a registered EC2 instance
type EC2Instance struct {
	Name         string            `json:"name"`
	InstanceID   string            `json:"instance_id"`
	State        string            `json:"state"`
	InstanceType string            `json:"instance_type"`
	PublicIP     string            `json:"public_ip,omitempty"`
	PrivateIP    string            `json:"private_ip,omitempty"`
	SSHKeyPath   string            `json:"ssh_key_path,omitempty"`
	SSHUser      string            `json:"ssh_user,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// NewEC2Instance creates a new EC2 instance record
func NewEC2Instance(name, instanceID, instanceType string) *EC2Instance {
	now := time.Now()
	return &EC2Instance{
		Name:         name,
		InstanceID:   instanceID,
		InstanceType: instanceType,
		State:        "unknown",
		CreatedAt:    now,
		UpdatedAt:    now,
		Tags:         make(map[string]string),
		Metadata:     make(map[string]string),
	}
}

// SetState updates the instance state
func (i *EC2Instance) SetState(state string) {
	i.State = state
	i.UpdatedAt = time.Now()
}

// SetIPs sets the IP addresses for the instance
func (i *EC2Instance) SetIPs(publicIP, privateIP string) {
	i.PublicIP = publicIP
	i.PrivateIP = privateIP
	i.UpdatedAt = time.Now()
}

// SetSSHConfig sets the SSH configuration for the instance
func (i *EC2Instance) SetSSHConfig(keyPath, user string) {
	i.SSHKeyPath = keyPath
	i.SSHUser = user
	i.UpdatedAt = time.Now()
}

// AddTag adds a tag to the instance
func (i *EC2Instance) AddTag(key, value string) {
	if i.Tags == nil {
		i.Tags = make(map[string]string)
	}
	i.Tags[key] = value
	i.UpdatedAt = time.Now()
}

// SetMetadata sets metadata for the instance
func (i *EC2Instance) SetMetadata(key, value string) {
	if i.Metadata == nil {
		i.Metadata = make(map[string]string)
	}
	i.Metadata[key] = value
	i.UpdatedAt = time.Now()
}

// IsRunning checks if the instance is running
func (i *EC2Instance) IsRunning() bool {
	return i.State == "running"
}

// IsStopped checks if the instance is stopped
func (i *EC2Instance) IsStopped() bool {
	return i.State == "stopped"
}

// S3Bucket represents an S3 bucket configuration
type S3Bucket struct {
	Name         string            `json:"name"`
	Region       string            `json:"region"`
	CreatedAt    time.Time         `json:"created_at"`
	LastAccessed time.Time         `json:"last_accessed,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// NewS3Bucket creates a new S3 bucket record
func NewS3Bucket(name, region string) *S3Bucket {
	now := time.Now()
	return &S3Bucket{
		Name:      name,
		Region:    region,
		CreatedAt: now,
		Tags:      make(map[string]string),
		Metadata:  make(map[string]string),
	}
}

// UpdateLastAccessed updates the last accessed time
func (b *S3Bucket) UpdateLastAccessed() {
	b.LastAccessed = time.Now()
}

// AddTag adds a tag to the bucket
func (b *S3Bucket) AddTag(key, value string) {
	if b.Tags == nil {
		b.Tags = make(map[string]string)
	}
	b.Tags[key] = value
}

// SetMetadata sets metadata for the bucket
func (b *S3Bucket) SetMetadata(key, value string) {
	if b.Metadata == nil {
		b.Metadata = make(map[string]string)
	}
	b.Metadata[key] = value
}
