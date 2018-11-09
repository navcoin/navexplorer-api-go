package config

import (
	"github.com/spf13/viper"
	"log"
	"sync"
)

type Config struct {
	Database struct{
		Host string
		Name string
	}

	Server struct{
		Port string
	}
}

var instance *Config
var once sync.Once

func Init(env string) *Config {
	once.Do(func() {
		viper.SetConfigName("config."+env)
		viper.AddConfigPath(".")

		instance = &Config{}

		if err := viper.ReadInConfig(); err != nil {
			log.Fatal(err)
		}

		if err := viper.Unmarshal(instance); err != nil {
			log.Fatal(err)
		}
	})

	return instance
}

func Get() *Config {
	if instance == nil {
		panic("Configuration has not been initialized")
	}

	return instance
}
