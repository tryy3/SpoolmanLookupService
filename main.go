package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tryy3/SpoolmanLookupService/config"
	"github.com/tryy3/SpoolmanLookupService/kafka"
	"github.com/tryy3/SpoolmanLookupService/models"
	"github.com/tryy3/SpoolmanLookupService/spoolman"
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

	spoolmanClient := spoolman.NewSpoolmanClient(cfg.SpoolmanAPIURL)
	printerService := spoolman.NewPrinterService()

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

		var spoolTransferInitiationEvent models.SpoolTransferInitiationEvent
		if err := json.Unmarshal(msg.Value, &spoolTransferInitiationEvent); err != nil {
			// If the message is completely wrong then we log it, commit and continue
			log.Printf("Error unmarshalling message: %v", err)
			if err := consumer.CommitMessages(ctx, msg); err != nil {
				log.Printf("Error committing message: %v", err)
			}
			continue
		}
		log.Printf("Unmarshalled message: %+v", spoolTransferInitiationEvent)

		spoolData, err := spoolmanClient.GetSpoolData(spoolTransferInitiationEvent.SpoolId)
		if err != nil {
			// TODO: Improve this error handling, we should retry if there is an actual error, but if the spool wasn't found then we should commit and continue
			log.Printf("Error getting spool data: %v", err)
			if err := consumer.CommitMessages(ctx, msg); err != nil {
				log.Printf("Error committing message: %v", err)
			}
			continue
		}
		log.Printf("Got spool data: %+v", spoolData)

		printerData, err := printerService.GetPrinterData(spoolTransferInitiationEvent.LocationId)
		if err != nil && err != spoolman.ErrPrinterNotFound {
			
			log.Printf("Error getting printer data: %v", err)
			continue
		}
		log.Printf("Got printer data: %+v", printerData)

		// Build the ready event - embed the initiation event to inherit all its fields
		readyEvent := models.SpoolTransferReadyEvent{
			SpoolTransferInitiationEvent: spoolTransferInitiationEvent,
			Filament:                     buildFilamentFromSpoolData(spoolData),
			Location:                     buildLocation(spoolTransferInitiationEvent.LocationId),
		}

		if err == nil {
			// Printer was found
			readyEvent.Type = models.EventTypePrinter
			readyEvent.Printer = &printerData
		} else {
			// No printer found (ErrPrinterNotFound), it's an inventory event
			readyEvent.Type = models.EventTypeInventory
		}

		outputValue, err := json.Marshal(readyEvent)
		if err != nil {
			log.Printf("Error marshalling ready event: %v", err)
			continue
		}

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

func buildFilamentFromSpoolData(spoolData models.SpoolmanSpoolData) models.Filament {
	return models.Filament{
		ID:              spoolData.Filament.Id,
		Name:            spoolData.Filament.Name,
		Material:        spoolData.Filament.Material,
		Color:           spoolData.Filament.ColorHex,
		Diameter:        spoolData.Filament.Diameter,
		Density:         spoolData.Filament.Density,
		RemainingWeight: spoolData.RemainingWeight,
		InitialWeight:   spoolData.InitialWeight,
		SpoolWeight:     spoolData.SpoolWeight,
		Temperature: models.FilamentTemperature{
			Nozzle: spoolData.Filament.ExtruderTemperature,
			Bed:    spoolData.Filament.BedTemperature,
		},
	}
}

func buildLocation(locationId string) models.Location {
	return models.Location{
		ID:   locationId,
		Name: locationId, // TODO: Look up actual location name if available
	}
}