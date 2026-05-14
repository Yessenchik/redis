package main

import (
	"context"
	"log"
	"net"

	pb "github.com/Yessenchik/generated/orderpayment"
	"github.com/Yessenchik/payment-service/config"
	deliverygrpc "github.com/Yessenchik/payment-service/delivery/grpc"
	pgrepo "github.com/Yessenchik/payment-service/repository/postgres"
	"github.com/Yessenchik/payment-service/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	dbpool, err := pgxpool.New(context.Background(), cfg.PostgresDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()

	lis, err := net.Listen("tcp", ":"+cfg.PaymentGRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	repo := pgrepo.NewPaymentRepository(dbpool)
	uc := usecase.NewPaymentUsecase(repo)

	server := grpc.NewServer()
	pb.RegisterPaymentServiceServer(server, deliverygrpc.NewPaymentServer(uc))

	log.Println("Payment gRPC server running on :" + cfg.PaymentGRPCPort)
	log.Fatal(server.Serve(lis))
}
