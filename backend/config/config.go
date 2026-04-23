package config

import (
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Env string

const (
	EnvLocal      Env = "local"
	EnvDevelopment Env = "development"
	EnvStaging    Env = "staging"
	EnvProduction Env = "production"
)

type Config struct {
	Env      Env            `mapstructure:"env"`
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Mongo    MongoConfig    `mapstructure:"mongo"`
	Neo4j    Neo4jConfig    `mapstructure:"neo4j"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	DB       string `mapstructure:"db"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

func (c PostgresConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" dbname=" + c.DB +
		" user=" + c.User +
		" password=" + c.Password +
		" sslmode=" + c.SSLMode
}

type MongoConfig struct {
	URI string `mapstructure:"uri"`
	DB  string `mapstructure:"db"`
}

type Neo4jConfig struct {
	URI      string `mapstructure:"uri"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret    string        `mapstructure:"secret"`
	AccessTTL time.Duration `mapstructure:"access_ttl"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

// Load reads configuration from embedded defaults, then overlays environment variables.
// Env vars map to config keys using double-underscore as separator
// (e.g. POSTGRES__HOST overrides postgres.host).
func Load() *Config {
	v := viper.New()

	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(defaultConfig)); err != nil {
		panic("failed to read default config: " + err.Error())
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))

	cfg := &Config{}
	if err := v.Unmarshal(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "mapstructure"
	}); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}

	return cfg
}
