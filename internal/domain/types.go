package domain

import "time"

type TenantID string

type User struct {
	ID        string
	TenantID  TenantID
	Username  string
	Status    string
	CreatedAt time.Time
}

type Session struct {
	ID        string
	TenantID  TenantID
	UserID    string
	JWTID     string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type MFADevice struct {
	ID                string
	TenantID          TenantID
	UserID            string
	DeviceName        string
	PublicKey         string
	State             string
	TrustPinEnrollID  string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type MFAChallenge struct {
	ID               string
	TenantID         TenantID
	UserID           string
	DeviceID         string
	Action           string
	State            string
	TrustPinChallengeID string
	IssuedAt         time.Time
	ExpiresAt        time.Time
	UpdatedAt        time.Time
}

type AuditLog struct {
	ID        string
	TenantID  TenantID
	UserID    *string
	EventType string
	Payload   string
	CreatedAt time.Time
}
