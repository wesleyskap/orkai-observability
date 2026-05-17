package observability

import "errors"

// Config defines the configuration settings for the observability package.
type Config struct {
	ServiceName string
	Environment string
	LogLevel    string
}

// ValidateConfig verifies that the configuration fields are not empty.
func ValidateConfig(cfg Config) error {
	if cfg.ServiceName == "" {
		return errors.New("ServiceName cannot be empty")
	}
	if cfg.Environment == "" {
		return errors.New("Environment cannot be empty")
	}
	return nil
}
