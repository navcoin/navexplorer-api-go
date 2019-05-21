package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type Config struct {
	Debug bool `yaml:"debug"`
	Ssl   bool

	Sentry struct {
		Active bool
		DSN    string
	}

	Server struct {
		Port   string
		Domain string
	}

	ElasticSearch   ElasticSearch `yaml:"elasticSearch"`
	Networks        []Network     `yaml:"networks"`

	SelectedNetwork string
}

type ElasticSearch struct {
	Urls        string `yaml:"urls"`
	Sniff       bool   `yaml:"sniff"`
	HealthCheck bool   `yaml:"healthCheck"`
}

type Network struct {
	Name string

	Host     string
	Port     int
	Username string
	Password string


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

		configFile, err := ioutil.ReadFile(fmt.Sprintf("./config.%s.yaml", env()))
		if err != nil {
			log.Fatal(err)
		}

		instance = &Config{}
		err = yaml.Unmarshal(configFile, instance)
		if err != nil {
			log.Fatal(err)
		}
	})
	return instance
}

func SelectNetwork(network string) {
	instance.SelectedNetwork = network
}

func env() string {
	var env = "prod"
	if len(os.Args) > 1 {
		env = os.Args[1]
	}
	log.Print("Environment: " + env)

	return env
}