package main

import (
	"fmt"
	"log"
	"os"

	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/spf13/cobra"
)

var (
	baseURL  string
	realm    string
	clientID string
	secret   string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "moh-sso",
		Short: "MOH SSO Management CLI",
		Long:  "CLI tool for managing Keycloak realms, clients, and users for the MOH SSO Dashboard",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if baseURL == "" || realm == "" || clientID == "" || secret == "" {
				return fmt.Errorf("please provide all Keycloak connection details (base-url, realm, client-id, secret)")
			}
			return nil
		},
	}

	// global flags
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", os.Getenv("KEYCLOAK_BASE_URL"), "Keycloak base URL")
	rootCmd.PersistentFlags().StringVar(&realm, "realm", os.Getenv("KEYCLOAK_REALM"), "Keycloak realm name")
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", os.Getenv("KEYCLOAK_CLIENT_ID"), "Keycloak client ID")
	rootCmd.PersistentFlags().StringVar(&secret, "secret", os.Getenv("KEYCLOAK_CLIENT_SECRET"), "Keycloak client secret")

	// add commands
	rootCmd.AddCommand(initRealmCmd)
	rootCmd.AddCommand(createClientCmd)
	rootCmd.AddCommand(listClientsCmd)
	rootCmd.AddCommand(createUserCmd)
	rootCmd.AddCommand(listUsersCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}
}

// --- Commands ---

// Initialize Realm
var initRealmCmd = &cobra.Command{
	Use:   "init-realm",
	Short: "Initialize the default Keycloak realm",
	Run: func(cmd *cobra.Command, args []string) {
		kc := keycloak.NewClient(baseURL, realm, clientID, secret)
		if err := kc.Authenticate(); err != nil {
			log.Fatalf("❌ Auth failed: %v", err)
		}
		if err := kc.EnsureRealmExists("moh-realm"); err != nil {
			log.Fatalf("❌ Failed to initialize realm: %v", err)
		}
		fmt.Println("✅ Realm 'moh-realm' is ready.")
	},
}

// --- Client Commands ---
var (
	clientName    string
	redirectUri   string
	clientBaseURL string
)

var createClientCmd = &cobra.Command{
	Use:   "create-client",
	Short: "Create a new client in Keycloak",
	Run: func(cmd *cobra.Command, args []string) {
		kc := keycloak.NewClient(baseURL, realm, clientID, secret)

		if err := kc.Authenticate(); err != nil {
			log.Fatalf("❌ Auth failed: %v", err)
		}

		params := keycloak.CreateClientParams{
			ClientID:               clientName,
			Name:                   clientName,
			Description:            "Created via CLI",
			BaseURL:                clientBaseURL,
			RootURL:                clientBaseURL,
			RedirectURIs:           []string{redirectUri},
			WebOrigins:             []string{clientBaseURL},
			PublicClient:           true,
			Protocol:               "openid-connect",
			StandardFlowEnabled:    true,
			ImplicitFlowEnabled:    false,
			DirectAccessGrants:     false,
			ServiceAccountsEnabled: false,
			Enabled:                true,
		}

		if err := kc.CreateClient(params); err != nil {
			log.Fatalf("❌ Error creating client: %v", err)
		}

		fmt.Printf("✅ Client '%s' created successfully\n", clientName)
	},
}

var listClientsCmd = &cobra.Command{
	Use:   "list-clients",
	Short: "List all Keycloak clients",
	Run: func(cmd *cobra.Command, args []string) {
		kc := keycloak.NewClient(baseURL, realm, clientID, secret)
		if err := kc.Authenticate(); err != nil {
			log.Fatalf("❌ Auth failed: %v", err)
		}
		clients, err := kc.ListClients()
		if err != nil {
			log.Fatalf("❌ Error fetching clients: %v", err)
		}
		for _, c := range clients {
			fmt.Printf("• %s (%s)\n", c.ClientID, c.ID)
		}
	},
}

// --- User Commands ---
var (
	username string
	password string
	role     string
)

var createUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Create a new Keycloak user",
	Run: func(cmd *cobra.Command, args []string) {
		kc := keycloak.NewClient(baseURL, realm, clientID, secret)
		if err := kc.Authenticate(); err != nil {
			log.Fatalf("❌ Auth failed: %v", err)
		}
		if err := kc.CreateUser( user model.User); err != nil {
			log.Fatalf("❌ Failed to create user: %v", err)
		}
		fmt.Printf("✅ User '%s' created successfully\n", username)
	},
}

var listUsersCmd = &cobra.Command{
	Use:   "list-users",
	Short: "List all Keycloak users",
	Run: func(cmd *cobra.Command, args []string) {
		kc := keycloak.NewClient(baseURL, realm, clientID, secret)
		if err := kc.Authenticate(); err != nil {
			log.Fatalf("❌ Auth failed: %v", err)
		}
		users, err := kc.ListUsers()
		if err != nil {
			log.Fatalf("❌ Error fetching users: %v", err)
		}
		for _, u := range users {
			fmt.Printf("• %s (%s)\n", u.Username, u.ID)
		}
	},
}

func init() {
	createClientCmd.Flags().StringVar(&clientName, "name", "", "Client name")
	createClientCmd.Flags().StringVar(&redirectUri, "redirect-uri", "", "Redirect URI")
	createClientCmd.Flags().StringVar(&clientBaseURL, "base-url", "", "Base URL")
	_ = createClientCmd.MarkFlagRequired("name")

	createUserCmd.Flags().StringVar(&username, "username", "", "Username")
	createUserCmd.Flags().StringVar(&password, "password", "", "Password")
	createUserCmd.Flags().StringVar(&role, "role", "user", "User role (default: user)")
	_ = createUserCmd.MarkFlagRequired("username")
	_ = createUserCmd.MarkFlagRequired("password")
}
