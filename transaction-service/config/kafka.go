package config

import (
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

var KafkaWriter *kafka.Writer

const TransferTopic = "transfer-requests"

func ConnectKafka() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}

	KafkaWriter = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaBroker),
		Topic:                  TransferTopic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		Async:                  true, // Mode Asinkron untuk peforma maksimal
		BatchSize:              100,  // Kirim dalam kelompok kecil agar efisien
		RequiredAcks:           kafka.RequireOne, // Cukup satu broker yang ACK agar cepat
	}
	log.Printf("[transaction-service] Kafka producer terhubung ke %s (topic: %s)\n", kafkaBroker, TransferTopic)
}
