package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestCoreProfilePreservesMandatoryValidationAndSkipsOptionalProviders(t *testing.T) {
	c := privateHTTPConfig()
	c.AllowInsecureHTTP = true
	c.PlatformProfile = ProfileCore
	c.DBHost = "primary"
	c.DBPort = "5432"
	c.DBUser = "core"
	c.DBPassword = "fixture"
	c.DBName = "core"
	c.StorageProvider = "unconfigured-provider"
	c.SMS.Enabled = true
	c.SMS.Provider = "unconfigured-provider"
	if err := validateConfig(c); err != nil {
		t.Fatal("core requires optional provider", err)
	}
	c.DBPassword = ""
	if err := validateConfig(c); err == nil || !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Fatal("primary credential validation lost", err)
	}
	c.DBPassword = "fixture"
	c.PlatformProfile = ProfileFull
	if err := validateConfig(c); err == nil {
		t.Fatal("full profile accepted absent optional providers")
	}
	c.PlatformProfile = "typo"
	if err := validateConfig(c); err == nil || !strings.Contains(err.Error(), "PLATFORM_PROFILE") {
		t.Fatal(err)
	}
}
func TestProfileEnvironmentAndDisablement(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("PLATFORM_PROFILE", " CORE ")
	setDefaults()
	viper.AutomaticEnv()
	if err := viper.BindEnv("PLATFORM_PROFILE", "PLATFORM_PROFILE"); err != nil {
		t.Fatal(err)
	}
	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		t.Fatal(err)
	}
	normalizeConfig(&c)
	if !c.CoreOnly() {
		t.Fatal("profile environment not bound")
	}
	for _, id := range OptionalModuleIDs() {
		if c.ModuleEnabled(id) {
			t.Fatal("core enabled", id)
		}
	}
	c.DisabledModules = "auth"
	if err := validateProfile(&c); err == nil {
		t.Fatal("mandatory disablement silently ignored")
	}
	c.DisabledModules = ""
	c.PlatformProfile = ""
	normalizeConfig(&c)
	if c.PlatformProfile != ProfileFull || !c.ModuleEnabled("metrics") {
		t.Fatal("default changed")
	}
	c.DisabledModules = " metrics ,audit"
	if c.ModuleEnabled("metrics") || c.ModuleEnabled("audit") || !c.ModuleEnabled("geojson") {
		t.Fatal("explicit disablement ignored")
	}
	if c.ModuleEnabled("unknown") {
		t.Fatal("unknown module enabled")
	}
}
