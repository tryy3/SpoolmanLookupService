package config

import (
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers      []string
	KafkaConsumerTopic string
	KafkaProducerTopic string
	KafkaConsumerGroup string
}

func Load() *Config {
	return &Config{
		KafkaBrokers:       getBrokers(),
		KafkaConsumerTopic: getEnv("KAFKA_CONSUMER_TOPIC", "3dprinter-filament-transfer-initiated"),
		KafkaProducerTopic: getEnv("KAFKA_PRODUCER_TOPIC", "output-topic"),
		KafkaConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "spoolman-lookup-service"),
	}
}

func getBrokers() []string {
	brokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	return strings.Split(brokers, ",")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

