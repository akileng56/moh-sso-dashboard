package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID                      string   `json:"id,omitempty"`
	ClientID                string   `json:"clientId"`
	Name                    string   `json:"name,omitempty"`
	Description             string   `json:"description,omitempty"`
	RootURL                 string   `json:"rootUrl,omitempty"`
	BaseURL                 string   `json:"baseUrl,omitempty"`
	AdminURL                string   `json:"adminUrl,omitempty"`
	SurrogateAuthRequired   bool     `json:"surrogateAuthRequired,omitempty"`
	Enabled                 bool     `json:"enabled"`
	AlwaysDisplayInConsole  bool     `json:"alwaysDisplayInConsole,omitempty"`
	ClientAuthenticatorType string   `json:"clientAuthenticatorType,omitempty"`
	RedirectUris            []string `json:"redirectUris,omitempty"`
	WebOrigins              []string `json:"webOrigins,omitempty"`
	NotBefore               int      `json:"notBefore,omitempty"`

	BearerOnly         bool `json:"bearerOnly,omitempty"` // Added omitempty for boolean
	ConsentRequired    bool `json:"consentRequired,omitempty"`
	StandardFlow       bool `json:"standardFlowEnabled,omitempty"`
	ImplicitFlow       bool `json:"implicitFlowEnabled,omitempty"`
	DirectAccess       bool `json:"directAccessGrantsEnabled,omitempty"`
	ServiceAccounts    bool `json:"serviceAccountsEnabled,omitempty"`
	PublicClient       bool `json:"publicClient,omitempty"`
	FrontChannelLogout bool `json:"frontchannelLogout,omitempty"`

	Protocol string `json:"protocol,omitempty"` // Added omitempty

	FullScopeAllowed          bool     `json:"fullScopeAllowed,omitempty"`
	NodeReRegistrationTimeout int      `json:"nodeReRegistrationTimeout,omitempty"`
	DefaultClientScopes       []string `json:"defaultClientScopes,omitempty"`
	OptionalClientScopes      []string `json:"optionalClientScopes,omitempty"`

	Access map[string]bool `json:"access,omitempty"` // Access/permissions flags

	// --- CRITICAL MISSING FIELDS ---

	// **Authentication/Security**
	Secret                  string `json:"secret,omitempty"`                  // The client secret for confidential clients
	ClientTemplate          string `json:"clientTemplate,omitempty"`          // The ID of the client template
	UseTemplateConfig       bool   `json:"useTemplateConfig,omitempty"`       // Use client template configuration (Deprecated)
	RootClientRealm         string `json:"rootClientRealm,omitempty"`         // Realm where client lives
	RegistrationAccessToken string `json:"registrationAccessToken,omitempty"` // For dynamic client registration

	// **Authorization Services (Permissions/Resources)**
	AuthorizationServicesEnabled bool `json:"authorizationServicesEnabled,omitempty"`

	// **Protocol Mappers & Roles**
	ProtocolMappers []ProtocolMapper `json:"protocolMappers,omitempty"` // Defines how claims map to tokens
	DefaultRoles    []string         `json:"defaultRoles,omitempty"`    // Default realm roles for service accounts

	// **Token/Session Configuration**
	AuthorizationURL string `json:"authorizationUrl,omitempty"` // SAML-specific
	TlsRequired      string `json:"tlsRequired,omitempty"`      // Options: 'none', 'all'

	// **Misc**
	Attributes                         map[string]string `json:"attributes,omitempty"`                         // Key/value map for custom settings (already in your struct, but repeated for context)
	AuthenticationFlowBindingOverrides map[string]string `json:"authenticationFlowBindingOverrides,omitempty"` // e.g., default: 'browser'

	// **Registered Nodes/Clustering**
	RegisteredNodes map[string]int `json:"registeredNodes,omitempty"` // Map of hostnames to registration time
}

// --- Auxiliary Struct for ProtocolMappers ---

// ProtocolMapper is a placeholder for the more complex Keycloak ProtocolMapperRepresentation
// which is used inside the Client object to define how attributes are mapped to tokens.
type ProtocolMapper struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Protocol        string            `json:"protocol,omitempty"`
	ProtocolMapper  string            `json:"protocolMapper,omitempty"`
	ConsentRequired bool              `json:"consentRequired,omitempty"`
	ConsentText     string            `json:"consentText,omitempty"`
	Config          map[string]string `json:"config,omitempty"` // Critical field for mapper configuration
}

type User struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FirstName        string              `json:"first_name,omitempty"`
	LastName         string              `json:"last_name,omitempty"`
	FullName         string              `json:"full_name,omitempty"`
	RealmRoles       []string            `json:"realm_roles"`
	IsAdmin          bool                `json:"is_admin"`
	ClientRoles      map[string][]string `json:"client_roles,omitempty"`
	Enabled          bool                `json:"enabled"`
	EmailVerified    bool                `json:"email_verified"`
	RequirePwdChange bool                `json:"require_pwd_change"`
	LastLoginAt      *time.Time          `json:"last_login_at,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	CreatedBy        string              `json:"created_by,omitempty"`
	UpdatedBy        string              `json:"updated_by,omitempty"`
}

type UserResponse struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email,omitempty"`
	FullName         string              `json:"full_name"`
	IsAdmin          bool                `json:"is_admin"`
	RealmRoles       []string            `json:"realm_roles"`
	ClientRoles      map[string][]string `json:"client_roles,omitempty"`
	IsActive         bool                `json:"is_active"`
	EmailVerified    bool                `json:"email_verified,omitempty"`
	RequirePwdChange bool                `json:"require_pwd_change,omitempty"`
	LastLoginAt      *time.Time          `json:"last_login_at,omitempty"`
	CreatedAt        *time.Time          `json:"created_at,omitempty"`
}

type ImportUserRow struct {
	RowNumber int      `json:"rowNumber"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Roles     []string `json:"roles"`
	Enabled   bool     `json:"enabled"`
	ClientIDs []string `json:"clientIds"`
	Errors    []string `json:"errors,omitempty"`
	Status    string   `json:"status,omitempty"`   // valid|invalid|success|failed|skipped
	ErrorMsg  string   `json:"errorMsg,omitempty"` // for failed
}

type PreviewResponse struct {
	JobID    string          `json:"jobId"`
	FileName string          `json:"fileName"`
	Total    int             `json:"total"`
	Valid    int             `json:"valid"`
	Invalid  int             `json:"invalid"`
	Rows     []ImportUserRow `json:"rows"` // include both valid + invalid for preview table
}

type ExecuteResponse struct {
	JobID        string `json:"jobId"`
	Total        int    `json:"total"`
	SuccessCount int    `json:"successCount"`
	FailureCount int    `json:"failureCount"`
	Status       string `json:"status"`
}

type JobStatusResponse struct {
	JobID        string          `json:"jobId"`
	FileName     string          `json:"fileName"`
	Status       string          `json:"status"`
	Total        int             `json:"total"`
	Valid        int             `json:"valid"`
	SuccessCount int             `json:"successCount"`
	FailureCount int             `json:"failureCount"`
	Rows         []ImportUserRow `json:"rows"`
}

type parsedRow struct {
	row ImportUserRow
	raw struct {
		clientIDs string
	}
}

type Notification struct {
	ID         uuid.UUID       `json:"id"`
	Type       string          `json:"type"`
	Title      string          `json:"title"`
	Message    string          `json:"message"`
	Severity   string          `json:"severity"`    // info | warning | critical
	TargetRole string          `json:"target_role"` // admin | super_admin | etc
	Metadata   json.RawMessage `json:"metadata"`    // JSONB from Postgres
	Read       bool            `json:"read"`
	CreatedAt  time.Time       `json:"created_at"`
}

type NotificationType string

// =====================================================
// AUTH / SECURITY
// =====================================================
const (
	LoginFailed            NotificationType = "LOGIN_FAILED"
	LoginSucceeded         NotificationType = "LOGIN_SUCCEEDED"
	SuspiciousLogin        NotificationType = "SUSPICIOUS_LOGIN"
	PasswordResetRequested NotificationType = "PASSWORD_RESET_REQUESTED"
	PasswordResetCompleted NotificationType = "PASSWORD_RESET_COMPLETED"
	AccountLocked          NotificationType = "ACCOUNT_LOCKED"
	AccountUnlocked        NotificationType = "ACCOUNT_UNLOCKED"
)

// =====================================================
// USERS
// =====================================================
const (
	UserImported    NotificationType = "USER_IMPORTED"
	UserCreated     NotificationType = "USER_CREATED"
	UserUpdated     NotificationType = "USER_UPDATED"
	UserDisabled    NotificationType = "USER_DISABLED"
	UserEnabled     NotificationType = "USER_ENABLED"
	UserRoleChanged NotificationType = "USER_ROLE_CHANGED"
)

// =====================================================
// CLIENTS / APPLICATIONS
// =====================================================
const (
	ClientCreated       NotificationType = "CLIENT_CREATED"
	ClientUpdated       NotificationType = "CLIENT_UPDATED"
	ClientDisabled      NotificationType = "CLIENT_DISABLED"
	ClientEnabled       NotificationType = "CLIENT_ENABLED"
	ClientSecretRotated NotificationType = "CLIENT_SECRET_ROTATED"
)

// =====================================================
// SYSTEM / OPERATIONS
// =====================================================
const (
	SystemStartup   NotificationType = "SYSTEM_STARTUP"
	SystemShutdown  NotificationType = "SYSTEM_SHUTDOWN"
	ConfigChanged   NotificationType = "CONFIG_CHANGED"
	BackupCompleted NotificationType = "BACKUP_COMPLETED"
	BackupFailed    NotificationType = "BACKUP_FAILED"
)

// =====================================================
// IMPORTS / BACKGROUND JOBS
// =====================================================
const (
	ImportStarted   NotificationType = "IMPORT_STARTED"
	ImportCompleted NotificationType = "IMPORT_COMPLETED"
	ImportFailed    NotificationType = "IMPORT_FAILED"
)

// =====================================================
// AUDIT / COMPLIANCE
// =====================================================
const (
	AuditExported   NotificationType = "AUDIT_EXPORTED"
	PolicyViolation NotificationType = "POLICY_VIOLATION"
)

func (t NotificationType) Severity() string {
	switch t {

	// 🔴 Critical
	case SuspiciousLogin,
		AccountLocked,
		ClientDisabled,
		ImportFailed,
		BackupFailed,
		PolicyViolation:
		return "critical"

	// 🟠 Warning
	case LoginFailed,
		UserDisabled,
		UserRoleChanged,
		ConfigChanged:
		return "warning"

	// ⚪ Info
	default:
		return "info"
	}
}

func (t NotificationType) Title() string {
	switch t {
	case LoginFailed:
		return "Failed login attempt"
	case SuspiciousLogin:
		return "Suspicious login detected"
	case UserImported:
		return "Users imported"
	case UserDisabled:
		return "User account disabled"
	case ClientDisabled:
		return "Client application disabled"
	case ImportFailed:
		return "Import job failed"
	default:
		return "System notification"
	}
}
