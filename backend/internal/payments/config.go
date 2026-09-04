package payments

import (
	"fmt"
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
)

type Config struct {
	SecretKey      string
	PublishableKey string
	WebhookSecret  string
}

func LoadConfig(app *pocketbase.PocketBase) (*Config, error) {
	raw, err := backendinternal.FindSettingData(app, "stripe")
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		SecretKey:      strings.TrimSpace(stringField(raw, "secret_key")),
		PublishableKey: strings.TrimSpace(stringField(raw, "publishable_key")),
		WebhookSecret:  strings.TrimSpace(stringField(raw, "webhook_secret")),
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("stripe secret_key not configured")
	}
	return cfg, nil
}

func stringField(data map[string]any, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}
