package config

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Astemirdum/library-service/pkg/auth0"

	"github.com/Astemirdum/library-service/pkg/kafka"

	"github.com/Astemirdum/library-service/pkg/logger"

	"github.com/kelseyhightower/envconfig"
)

type HTTPServer struct {
	Host         string        `yaml:"host" envconfig:"GATEWAY_HTTP_HOST"`
	Port         string        `yaml:"port" envconfig:"GATEWAY_HTTP_PORT"`
	ReadTimeout  time.Duration `yaml:"readTimeout" envconfig:"HTTP_READ"`
	WriteTimeout time.Duration
}

type RatingHTTPServer struct {
	Host string `envconfig:"RATING_HTTP_HOST"`
	Port string `envconfig:"RATING_HTTP_PORT"`
}

type LibraryHTTPServer struct {
	Host string `envconfig:"LIBRARY_HTTP_HOST"`
	Port string `envconfig:"LIBRARY_HTTP_PORT"`
}

type ReservationHTTPServer struct {
	Host string `envconfig:"RESERVATION_HTTP_HOST"`
	Port string `envconfig:"RESERVATION_HTTP_PORT"`
}

type Config struct {
	Server                HTTPServer `yaml:"server"`
	Kafka                 kafka.Config
	Auth0                 auth0.Config
	ReservationHTTPServer ReservationHTTPServer
	LibraryHTTPServer     LibraryHTTPServer
	RatingHTTPServer      RatingHTTPServer
	Log                   logger.Log `yaml:"log"`
}

var (
	once sync.Once
	cfg  Config
)

// NewConfig reads config from environment.
func NewConfig(ops ...Option) Config {
	once.Do(func() {
		var config Config
		for _, op := range ops {
			op(&config)
		}
		err := envconfig.Process("", &config)
		if err != nil {
			log.Fatal("NewConfig ", err)
		}
		cfg = config
		printConfig(cfg)
	})

	return cfg
}

func printConfig(cfg Config) {
	jscfg, _ := json.MarshalIndent(cfg, "", "	") //nolint:errcheck
	fmt.Println(string(jscfg))
}
