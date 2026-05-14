package main

import (
	"context"
	"log"

	pb "github.com/Yessenchik/generated/orderpayment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)

	resp, err := client.ListPayments(context.Background(), &pb.ListPaymentsRequest{
		Status: "",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range resp.Payments {
		log.Printf("id=%s order_id=%s amount=%.2f status=%s",
			p.Id, p.OrderId, p.Amount, p.Status)
	}
}
