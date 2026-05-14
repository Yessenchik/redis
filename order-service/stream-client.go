package main

import (
	"context"
	"io"
	"log"

	pb "github.com/Yessenchik/generated/orderpayment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	stream, err := client.SubscribeToOrderUpdates(context.Background(), &pb.OrderRequest{
		OrderId: "1",
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("order_id=%s status=%s updated_at=%s", msg.Status, msg.UpdatedAt.AsTime())
	}
}
