package mq

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type Config struct {
	uri       string
	exchange  string
	queueName string
	logger    *logrus.Entry
	handler   MessageHandler
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

func WithLogger(logger *logrus.Entry) Option {
	return func(config *Config) {
		config.logger = logger
	}
}

func generateConfig(opts ...Option) (*Config, error) {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}

	if config.uri == "" ||
		config.queueName == "" ||
		config.exchange == "" {
		return nil, fmt.Errorf("uri or queue name or exchange is empty")
	}

	if config.handler == nil {
		return nil, fmt.Errorf("message handler is nil")
	}

	if config.logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	return config, nil
}
