package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
}

type Server struct {
	Port            int           `yaml:"-"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type Database struct {
	SSLMode     string `yaml:"sslmode"`
	MaxConns    int32  `yaml:"max_connections"`
	MinConns    int32  `yaml:"min_connections"`
	MaxIdleTime int    `yaml:"max_idle_time"`
	MaxLifeTime int    `yaml:"max_lifetime"`

	User     string `yaml:"-"`
	Password string `yaml:"-"`
	Name     string `yaml:"-"`
	Host     string `yaml:"-"`
	Port     string `yaml:"-"`
}

func Load(configPath, envPath string) (*Config, error) {
	if err := godotenv.Load(envPath); err != nil {
		return nil, err
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)

	if err != nil {
		return nil, fmt.Errorf("invalid PORT value %q: %s", portStr, err.Error())
	}

	config.Server.Port = port
	config.Database.User = os.Getenv("POSTGRES_USER")
	config.Database.Password = os.Getenv("POSTGRES_PASSWORD")
	config.Database.Name = os.Getenv("POSTGRES_DB")
	config.Database.Host = os.Getenv("POSTGRES_HOST")
	config.Database.Port = os.Getenv("POSTGRES_PORT")

	return &config, nil
}

func (c *Config) Address() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

func (d *Database) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}
