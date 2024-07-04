package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env            string `yaml:"env" env-required:"true"`
	MigrationsPath string `yaml:"migrations_path" env-required:"true"`
	AppID          int32  `yaml:"app_id" env-required:"true"`
	Database       `yaml:"database" env-required:"true"`
	HTTPServer     `yaml:"http_server" env-required:"true"`
	SSOClient      `yaml:"sso_client" env-required:"true"`
}

type Database struct {
	Host     string        `yaml:"host" env-required:"true"`
	Port     int           `yaml:"port" env-required:"true"`
	User     string        `yaml:"user" env-required:"true"`
	Pass     string        `yaml:"pass" env-required:"true"`
	Name     string        `yaml:"name" env-required:"true"`
	SSLMode  string        `yaml:"sslmode" env-required:"true"`
	Timeout  time.Duration `yaml:"timeout" env-required:"true"`
	Delay    time.Duration `yaml:"delay" env-required:"true"`
	Attempts int           `yaml:"attempts" env-required:"true"`
}

type HTTPServer struct {
	Host        string        `yaml:"host" env-required:"true"`
	Port        int           `yaml:"port" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

type SSOClient struct {
	Address      string        `yaml:"address" env-required:"true"`
	Timeout      time.Duration `yaml:"timeout" env-required:"true"`
	RetriesCount int           `yaml:"retries_count" env-required:"true"`
}

func MustLoad() *Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("failed to load environment file, error: ", err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	cfg := MustLoadWithPath(configPath)
	return cfg
}

func MustLoadWithPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("config file not found")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal("failed to read config, error: ", err)
	}

	return &cfg
}
