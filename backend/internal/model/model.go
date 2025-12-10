package model

import (
	"time"
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
	ID          string
	Username    string
	FirstName   string
	LastName    string
	Email       string
	Enabled     bool
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt time.Time
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	IsActive bool   `json:"is_active"`
}
