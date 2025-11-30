package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tryy3/SpoolmanLookupService/config"
	"github.com/tryy3/SpoolmanLookupService/kafka"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting Spoolman Lookup Service")
	log.Printf("Kafka Brokers: %v", cfg.KafkaBrokers)
	log.Printf("Consumer Topic: %s", cfg.KafkaConsumerTopic)
	log.Printf("Producer Topic: %s", cfg.KafkaProducerTopic)
	log.Printf("Consumer Group: %s", cfg.KafkaConsumerGroup)

	// Create consumer and producer
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaConsumerTopic, cfg.KafkaConsumerGroup)
	defer consumer.Close()

	producer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaProducerTopic)
	defer producer.Close()

	// Setup context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping...")
		cancel()
	}()

	// Start consuming messages
	log.Println("Starting to consume messages...")
	for {
		msg, err := consumer.FetchMessage(ctx)
		log.Println("Fetched message from Kafka")
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled, shutting down
				break
			}
			log.Printf("Error fetching message: %v", err)
			continue
		}

		log.Printf("Received message: key=%s value=%s", string(msg.Key), string(msg.Value))

		// TODO: Process the message here
		// 1. Parse the incoming message
		// 2. Look up data from Spoolman API
		// 3. Create the output message

		// Example: produce a response message
		outputValue := processMessage(msg.Value)
		if err := producer.WriteMessage(ctx, msg.Key, outputValue); err != nil {
			log.Printf("Error producing message: %v", err)
			continue
		}

		// Commit the message after successful processing
		if err := consumer.CommitMessages(ctx, msg); err != nil {
			log.Printf("Error committing message: %v", err)
		}

		log.Printf("Successfully processed and forwarded message")
	}

	log.Println("Service stopped")
}

// processMessage is a placeholder for your message processing logic.
// This is where you would:
// 1. Parse the incoming message
// 2. Call Spoolman API to get additional data
// 3. Create and return the enriched message
func processMessage(input []byte) []byte {
	// TODO: Implement your Spoolman lookup logic here
	// For now, just echo back the input
	return input
}
