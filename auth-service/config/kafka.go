package config

import (
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

var KafkaWriter *kafka.Writer

const AuditTopic = "audit-logs"

func ConnectKafka() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}

	KafkaWriter = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaBroker),
		Topic:                  AuditTopic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		Async:                  true, // Sangat penting untuk peforma
		RequiredAcks:           kafka.RequireNone, // Kita tidak butuh ACK untuk audit log demi kecepatan maksimal
	}
	log.Printf("[auth-service] Kafka producer (Audit) terhubung ke %s\n", kafkaBroker)
}
