package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"sync"
)

type Config struct {
	Server struct{
		Port string
	}

	Network string

	ElasticSearch struct {
		Urls string
		Sniff bool
		HealthCheck bool
	}

	CommunityFund struct {
		BlocksInCycle  int
		MinQuorum      float64
		ProposalVoting struct {
			Cycles int
			Accept float64
			Reject float64
		}
		PaymentVoting struct {
			Cycles int
			Accept float64
			Reject float64
		}
	}

	SoftFork struct {
		BlocksInCycle int
		Accept        float64
	}
}

var instance *Config
var once sync.Once

func Get() *Config {
	once.Do(func() {
		log.Println("Creating Config")
		var env = "dev"
		if len(os.Args) > 1 {
			env = os.Args[1]
		}

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