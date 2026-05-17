package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "capstone/proto/transaction"
	"transaction-service/config"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type TransactionServer struct {
	pb.UnimplementedTransactionServiceServer
}

func (s *TransactionServer) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	fromAccountID, errFrom := uuid.Parse(req.SenderId)
	toAccountID, errTo := uuid.Parse(req.TargetAccount)

	if errFrom != nil || errTo != nil {
		return &pb.TransferResponse{
			StatusCode: 400,
			Message:    "Invalid Account ID format",
		}, nil
	}

	if req.Amount <= 0 {
		return &pb.TransferResponse{
			StatusCode: 400,
			Message:    "Amount must be strictly positive",
		}, nil
	}
	if fromAccountID == toAccountID {
		return &pb.TransferResponse{
			StatusCode: 400,
			Message:    "Cannot transfer to same self-account",
		}, nil
	}

	referenceNumber := fmt.Sprintf("TRX-%d", time.Now().UnixNano())
	message := map[string]interface{}{
		"reference_number": referenceNumber,
		"from_account_id":  fromAccountID.String(),
		"to_account_id":    toAccountID.String(),
		"amount":           req.Amount,
		"submitted_at":     time.Now().UTC().Format(time.RFC3339),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return &pb.TransferResponse{
			StatusCode: 500,
			Message:    "Gagal menyiapkan pesan transaksi",
		}, nil
	}

	// Kirim ke Kafka
	err = config.KafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(referenceNumber),
		Value: messageBytes,
	})

	if err != nil {
		return &pb.TransferResponse{
			StatusCode: 503,
			Message:    "Sistem antrean sementara tidak tersedia, coba lagi",
		}, nil
	}

	return &pb.TransferResponse{
		StatusCode:    202,
		Message:       "Transfer sedang diproses di antrean.",
		TransactionId: referenceNumber,
	}, nil
}
