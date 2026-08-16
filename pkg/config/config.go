package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Context struct {
	URL                string            `mapstructure:"url" json:"url" yaml:"url"`
	APIToken           string            `mapstructure:"api_token,omitempty" json:"api_token,omitempty" yaml:"api_token,omitempty"`
	Username           string            `mapstructure:"username,omitempty" json:"username,omitempty" yaml:"username,omitempty"`
	Password           string            `mapstructure:"password,omitempty" json:"password,omitempty" yaml:"password,omitempty"`
	HTTPUser           string            `mapstructure:"http_user,omitempty" json:"http_user,omitempty" yaml:"http_user,omitempty"`
	HTTPPassword       string            `mapstructure:"http_password,omitempty" json:"http_password,omitempty" yaml:"http_password,omitempty"`
	HTTPHeaders        map[string]string `mapstructure:"http_headers,omitempty" json:"http_headers,omitempty" yaml:"http_headers,omitempty"`
	TLSCertFile        string            `mapstructure:"tls_cert_file,omitempty" json:"tls_cert_file,omitempty" yaml:"tls_cert_file,omitempty"`
	TLSKeyFile         string            `mapstructure:"tls_key_file,omitempty" json:"tls_key_file,omitempty" yaml:"tls_key_file,omitempty"`
	TLSCAFile          string            `mapstructure:"tls_ca_file,omitempty" json:"tls_ca_file,omitempty" yaml:"tls_ca_file,omitempty"`
	SafetyLevel        string            `mapstructure:"safety_level" json:"safety_level" yaml:"safety_level"`
	OutputFormat       string            `mapstructure:"output_format" json:"output_format" yaml:"output_format"`
	Timeout            int               `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
	InsecureSkipVerify bool              `mapstructure:"insecure_skip_verify" json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

type Config struct {
	ActiveContext string             `mapstructure:"active_context" json:"active_context" yaml:"active_context"`
	Contexts      map[string]Context `mapstructure:"contexts" json:"contexts" yaml:"contexts"`
}

func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zbxctl", "config.yaml"), nil
}

func DefaultConfig() *Config {
	return &Config{
		ActiveContext: "default",
		Contexts: map[string]Context{
			"default": {
				URL:          "http://localhost:8080/zabbix/api_jsonrpc.php",
				SafetyLevel:  "readonly",
				OutputFormat: "auto",
				Timeout:      30,
			},
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return nil, fmt.Errorf("failed to get default config path: %w", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file at %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config: %w", err)
	}

	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]Context)
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config, path string) error {
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get default config path: %w", err)
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to yaml: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", path, err)
	}

	return nil
}

func (c *Config) GetActiveContext() (*Context, string, error) {
	if c.ActiveContext == "" {
		return nil, "", fmt.Errorf("no active context configured")
	}

	ctx, ok := c.Contexts[c.ActiveContext]
	if !ok {
		return nil, c.ActiveContext, fmt.Errorf("active context %q not found in config", c.ActiveContext)
	}

	if ctx.SafetyLevel == "" {
		ctx.SafetyLevel = "readonly"
	}
	if ctx.OutputFormat == "" {
		ctx.OutputFormat = "auto"
	}
	if ctx.Timeout <= 0 {
		ctx.Timeout = 30
	}

	return &ctx, c.ActiveContext, nil
}

func (c Context) Redacted() Context {
	redacted := c
	if redacted.APIToken != "" {
		redacted.APIToken = "[REDACTED]"
	}
	if redacted.Password != "" {
		redacted.Password = "[REDACTED]"
	}
	if redacted.HTTPPassword != "" {
		redacted.HTTPPassword = "[REDACTED]"
	}
	if len(redacted.HTTPHeaders) > 0 {
		headers := make(map[string]string, len(redacted.HTTPHeaders))
		for k := range redacted.HTTPHeaders {
			headers[k] = "[REDACTED]"
		}
		redacted.HTTPHeaders = headers
	}
	return redacted
}

func (c *Config) Redacted() *Config {
	if c == nil {
		return nil
	}
	redacted := &Config{
		ActiveContext: c.ActiveContext,
		Contexts:      make(map[string]Context, len(c.Contexts)),
	}
	for name, ctx := range c.Contexts {
		redacted.Contexts[name] = ctx.Redacted()
	}
	return redacted
}

