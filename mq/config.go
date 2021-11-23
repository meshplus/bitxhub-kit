package mq

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
)

type Config struct {
	uri          string
	exchange     string
	exchangeType string
	queueName    string
	routingKey   string
	logger       hclog.Logger
	handler      MessageHandler
}

type Option func(*Config)

func WithURI(uri string) Option {
	return func(config *Config) {
		config.uri = uri
	}
}

func WithExchange(exchange string) Option {
	return func(config *Config) {
		config.exchange = exchange
	}
}

func WithExchangeType(exchangeType string) Option {
	return func(config *Config) {
		config.exchangeType = exchangeType
	}
}

func WithQueueName(queueName string) Option {
	return func(config *Config) {
		config.queueName = queueName
	}
}

func WithHandler(h MessageHandler) Option {
	return func(config *Config) {
		config.handler = h
	}
}

func WithLogger(logger hclog.Logger) Option {
	return func(config *Config) {
		config.logger = logger
	}
}

func WithRoutingKey(routingKey string) Option {
	return func(config *Config) {
		config.routingKey = routingKey
	}
}

func generateConfig(opts ...Option) (*Config, error) {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}

	if config.uri == "" ||
		config.queueName == "" ||
		config.exchange == "" || config.exchangeType == "" {
		return nil, fmt.Errorf("uri or queue name or exchange is empty")
	}

	if config.handler == nil {
		return nil, fmt.Errorf("message handler is nil")
	}

	if config.logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	if config.routingKey == "" {
		config.routingKey = "MQLog"
	}

	return config, nil
}
