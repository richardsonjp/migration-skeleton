package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

// Config global setting
var Config = struct {
	DB struct {
		Username string `env:"DB_USERNAME" envDefault:""`
		Password string `env:"DB_PASSWORD" envDefault:""`
		Name     string `env:"DB_NAME" envDefault:""`
		Host     string `env:"DB_HOST" envDefault:"localhost"`
		Port     string `env:"DB_PORT" envDefault:"5432"`
		SSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`
		Schemas  string `env:"DB_SCHEMAS" envDefault:"public"`
	}

	// GetSchemas returns the list of schemas as a slice
	GetSchemas func() []string `json:"-"`
}{}

// Init initialize config
func Init() {
	if Exists("./.env") {
		if err := godotenv.Load(); err != nil {
			panic(err)
		}
	}

	if err := env.Parse(&Config); err != nil {
		panic(err)
	}

	Config.GetSchemas = parseSchemas
}

// parseSchemas parses comma-separated schema list into a slice
func parseSchemas() []string {
	var schemas []string
	for _, s := range strings.Split(Config.DB.Schemas, ",") {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			schemas = append(schemas, trimmed)
		}
	}
	if len(schemas) == 0 {
		return []string{"public"}
	}
	return schemas
}

// Show for cli print setting
func Show() {
	Init()
	str, _ := json.MarshalIndent(Config, "", " ")
	fmt.Printf("Config: %s\n", str)
}

// Exists check file exist
func Exists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !(info.IsDir())
}
