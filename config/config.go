package config

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	DebugMode          bool
	AllowRegistrations bool
	TimezoneName       string
}

func envOrDefaultBool(key string, defaultValue bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return strings.TrimSpace(strings.ToLower(val)) == "true"
}

func envOrDefaultString(key string, defaultValue string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return strings.TrimSpace(val)
}

func Init() *Config {
	cfg := &Config{}
	cfg.DebugMode = envOrDefaultBool("DEBUG", false)
	cfg.AllowRegistrations = envOrDefaultBool("ALLOW_REGISTRATION", false)
	cfg.TimezoneName = envOrDefaultString("TZ", "UTC")

	flag.BoolVar(&cfg.DebugMode, "debug", cfg.DebugMode, "enable debug mode")
	flag.BoolVar(&cfg.AllowRegistrations, "allowRegistrations", cfg.AllowRegistrations, "allow registrations")
	flag.Parse()

	return cfg
}
