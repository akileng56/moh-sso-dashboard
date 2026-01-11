package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID                                 string            `json:"id"`
	ClientID                           string            `json:"clientId"`
	Name                               string            `json:"name,omitempty"`
	Description                        string            `json:"description,omitempty"`
	RootURL                            string            `json:"rootUrl,omitempty"`
	BaseURL                            string            `json:"baseUrl,omitempty"`
	AdminURL                           string            `json:"adminUrl,omitempty"`
	SurrogateAuthRequired              bool              `json:"surrogateAuthRequired,omitempty"`
	Enabled                            bool              `json:"enabled"`
	AlwaysDisplayInConsole             bool              `json:"alwaysDisplayInConsole,omitempty"`
	ClientAuthenticatorType            string            `json:"clientAuthenticatorType,omitempty"`
	RedirectUris                       []string          `json:"redirectUris,omitempty"`
	WebOrigins                         []string          `json:"webOrigins,omitempty"`
	NotBefore                          int               `json:"notBefore,omitempty"`
	BearerOnly                         bool              `json:"bearerOnly,omitempty"`
	ConsentRequired                    bool              `json:"consentRequired,omitempty"`
	StandardFlow                       bool              `json:"standardFlowEnabled,omitempty"`
	ImplicitFlow                       bool              `json:"implicitFlowEnabled,omitempty"`
	DirectAccess                       bool              `json:"directAccessGrantsEnabled,omitempty"`
	ServiceAccounts                    bool              `json:"serviceAccountsEnabled,omitempty"`
	PublicClient                       bool              `json:"publicClient,omitempty"`
	FrontChannelLogout                 bool              `json:"frontchannelLogout,omitempty"`
	Protocol                           string            `json:"protocol,omitempty"`
	FullScopeAllowed                   bool              `json:"fullScopeAllowed,omitempty"`
	NodeReRegistrationTimeout          int               `json:"nodeReRegistrationTimeout,omitempty"`
	DefaultClientScopes                []string          `json:"defaultClientScopes,omitempty"`
	OptionalClientScopes               []string          `json:"optionalClientScopes,omitempty"`
	Access                             map[string]bool   `json:"access,omitempty"`
	Secret                             string            `json:"secret,omitempty"`
	ClientTemplate                     string            `json:"clientTemplate,omitempty"`
	UseTemplateConfig                  bool              `json:"useTemplateConfig,omitempty"`
	RootClientRealm                    string            `json:"rootClientRealm,omitempty"`
	RegistrationAccessToken            string            `json:"registrationAccessToken,omitempty"`
	AuthorizationServicesEnabled       bool              `json:"authorizationServicesEnabled,omitempty"`
	ProtocolMappers                    []ProtocolMapper  `json:"protocolMappers,omitempty"`
	DefaultRoles                       []string          `json:"defaultRoles,omitempty"`
	AuthorizationURL                   string            `json:"authorizationUrl,omitempty"`
	TlsRequired                        string            `json:"tlsRequired,omitempty"`
	Attributes                         map[string]string `json:"attributes,omitempty"`
	AuthenticationFlowBindingOverrides map[string]string `json:"authenticationFlowBindingOverrides,omitempty"`
	RegisteredNodes                    map[string]int    `json:"registeredNodes,omitempty"`
}

type ProtocolMapper struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Protocol        string            `json:"protocol,omitempty"`
	ProtocolMapper  string            `json:"protocolMapper,omitempty"`
	ConsentRequired bool              `json:"consentRequired,omitempty"`
	ConsentText     string            `json:"consentText,omitempty"`
	Config          map[string]string `json:"config,omitempty"`
}

type User struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FirstName        string              `json:"firstName,omitempty"`
	LastName         string              `json:"lastName,omitempty"`
	FullName         string              `json:"fullName,omitempty"`
	RealmRoles       []string            `json:"realmRoles"`
	IsAdmin          bool                `json:"isAdmin"`
	ClientRoles      map[string][]string `json:"clientRoles,omitempty"`
	Enabled          bool                `json:"enabled"`
	EmailVerified    bool                `json:"emailVerified"`
	RequirePwdChange bool                `json:"requirePwdChange"`
	LastLoginAt      *time.Time          `json:"lastLoginAt,omitempty"`
	CreatedAt        time.Time           `json:"createdAt"`
	UpdatedAt        time.Time           `json:"updatedAt"`
	CreatedBy        string              `json:"createdBy,omitempty"`
	UpdatedBy        string              `json:"updatedBy,omitempty"`
}

type UserResponse struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email,omitempty"`
	FullName         string              `json:"fullName"`
	IsAdmin          bool                `json:"isAdmin"`
	RealmRoles       []string            `json:"realmRoles"`
	ClientRoles      map[string][]string `json:"clientRoles,omitempty"`
	IsActive         bool                `json:"isActive"`
	EmailVerified    bool                `json:"emailVerified,omitempty"`
	RequirePwdChange bool                `json:"requirePwdChange,omitempty"`
	LastLoginAt      *time.Time          `json:"lastLoginAt,omitempty"`
	CreatedAt        *time.Time          `json:"createdAt,omitempty"`
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
	Status    string   `json:"status,omitempty"`
	ErrorMsg  string   `json:"errorMsg,omitempty"`
}

type PreviewResponse struct {
	JobID    string          `json:"jobId"`
	FileName string          `json:"fileName"`
	Total    int             `json:"total"`
	Valid    int             `json:"valid"`
	Invalid  int             `json:"invalid"`
	Rows     []ImportUserRow `json:"rows"`
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

const (
	LoginFailed            NotificationType = "LOGIN_FAILED"
	TokenRefreshFailed     NotificationType = "TOKEN_REFRESH_FAILED"
	LoginSucceeded         NotificationType = "LOGIN_SUCCEEDED"
	SuspiciousLogin        NotificationType = "SUSPICIOUS_LOGIN"
	PasswordResetRequested NotificationType = "PASSWORD_RESET_REQUESTED"
	PasswordResetCompleted NotificationType = "PASSWORD_RESET_COMPLETED"
	AccountLocked          NotificationType = "ACCOUNT_LOCKED"
	AccountUnlocked        NotificationType = "ACCOUNT_UNLOCKED"
)
const (
	UserImported    NotificationType = "USER_IMPORTED"
	UserCreated     NotificationType = "USER_CREATED"
	UserUpdated     NotificationType = "USER_UPDATED"
	UserDisabled    NotificationType = "USER_DISABLED"
	UserDeleted     NotificationType = "USER_DELETED"
	UserEnabled     NotificationType = "USER_ENABLED"
	UserRoleChanged NotificationType = "USER_ROLE_CHANGED"
)
const (
	ClientCreated       NotificationType = "CLIENT_CREATED"
	ClientUpdated       NotificationType = "CLIENT_UPDATED"
	ClientDeleted       NotificationType = "CLIENT_DELETED"
	ClientDisabled      NotificationType = "CLIENT_DISABLED"
	ClientEnabled       NotificationType = "CLIENT_ENABLED"
	ClientSecretRotated NotificationType = "CLIENT_SECRET_ROTATED"
)
const (
	SystemStartup   NotificationType = "SYSTEM_STARTUP"
	SystemShutdown  NotificationType = "SYSTEM_SHUTDOWN"
	ConfigChanged   NotificationType = "CONFIG_CHANGED"
	BackupCompleted NotificationType = "BACKUP_COMPLETED"
	BackupFailed    NotificationType = "BACKUP_FAILED"
)

const (
	ImportStarted   NotificationType = "IMPORT_STARTED"
	ImportCompleted NotificationType = "IMPORT_COMPLETED"
	ImportFailed    NotificationType = "IMPORT_FAILED"
)

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
	case SystemStartup,
		BackupCompleted:
		return "info"

	// ⚪ Info
	default:
		return "info"
	}
}

func (t NotificationType) Title() string {
	switch t {
	case SystemStartup:
		return "System started"
	case SystemShutdown:
		return "System shutdown"
	case ConfigChanged:
		return "Configuration changed"
	case BackupCompleted:
		return "Backup completed"
	case BackupFailed:
		return "Backup failed"
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
