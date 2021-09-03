package config

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/log"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Logging        bool
	LogPath        string
	Debug          bool
	ElasticSearch  ElasticSearchConfig
	Index          map[string]string
	Server         ServerConfig
	Legacy         bool
	Subscribe      bool
	RabbitMq       RabbitMqConfig
	DefaultNetwork string
	User           string
	Password       string
}

type ElasticSearchConfig struct {
	Hosts       []string
	Sniff       bool
	HealthCheck bool
	Debug       bool
	Username    string
	Password    string
}

type ServerConfig struct {
	Port int
}

type RabbitMqConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Prefix   string
}

func Init() {
	err := godotenv.Load()
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("Unable to init config")
	}

	log.NewLogger(fmt.Sprintf("%s/indexer.log", Get().LogPath), Get().Debug)
}

func Get() *Config {
	return &Config{
		Logging: getBool("LOGGING", false),
		LogPath: getString("LOG_PATH", "/app/logs"),
		Debug:   getBool("DEBUG", false),
		ElasticSearch: ElasticSearchConfig{
			Hosts:       getSlice("ELASTIC_SEARCH_HOSTS", make([]string, 0), ","),
			Sniff:       getBool("ELASTIC_SEARCH_SNIFF", true),
			HealthCheck: getBool("ELASTIC_SEARCH_HEALTH_CHECK", true),
			Debug:       getBool("ELASTIC_SEARCH_DEBUG", false),
			Username:    getString("ELASTIC_SEARCH_USERNAME", ""),
			Password:    getString("ELASTIC_SEARCH_PASSWORD", ""),
		},
		Index: map[string]string{
			"devnet":  getString("INDEX_DEVNET", "v2"),
			"testnet": getString("INDEX_TESTNET", "v1"),
			"mainnet": getString("INDEX_MAINNET", "v2"),
		},
		Server: ServerConfig{
			Port: getInt("PORT", 8080),
		},
		Legacy:    getBool("LEGACY", true),
		Subscribe: getBool("SUBSCRIBE", false),
		RabbitMq: RabbitMqConfig{
			User:     getString("RABBITMQ_USER", "user"),
			Password: getString("RABBITMQ_PASSWORD", "user"),
			Host:     getString("RABBITMQ_HOST", ""),
			Port:     getInt("RABBITMQ_PORT", 5672),
			Prefix:   getString("RABBITMQ_PREFIX", os.Getenv("POD_NAME")),
		},
		DefaultNetwork: getString("DEFAULT_NETWORK", "mainnet"),
		User:           getString("AUTH_USER", "user"),
		Password:       getString("AUTH_PASSWORD", "password"),
	}
}

func Account() gin.Accounts {
	config := Get()
	return gin.Accounts{
		config.User: config.Password,
	}
}

func getString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

func getInt(key string, defaultValue int) int {
	valStr := getString(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}

	return defaultValue
}

func getBool(key string, defaultValue bool) bool {
	valStr := getString(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultValue
}

func getSlice(key string, defaultVal []string, sep string) []string {
	valStr := getString(key, "")
	if valStr == "" {
		return defaultVal
	}

	return strings.Split(valStr, sep)
}
