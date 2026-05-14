package main

import (
	"context"
	"log"
	"net"

	pb "github.com/Yessenchik/generated/orderpayment"
	"github.com/Yessenchik/order-service/cache"
	"github.com/Yessenchik/order-service/config"
	grpcdelivery "github.com/Yessenchik/order-service/delivery/grpc"
	httpdelivery "github.com/Yessenchik/order-service/delivery/http"
	pgrepo "github.com/Yessenchik/order-service/repository/postgres"
	"github.com/Yessenchik/order-service/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentGRPCClient struct {
	client pb.PaymentServiceClient
}

func (p *PaymentGRPCClient) ProcessPayment(ctx context.Context, orderID string, amount float64, customerEmail string) (bool, string, error) {
	resp, err := p.client.ProcessPayment(ctx, &pb.PaymentRequest{
		OrderId:       orderID,
		Amount:        amount,
		CustomerEmail: customerEmail,
	})
	if err != nil {
		return false, "", err
	}
	return resp.Success, resp.Message, nil
}

func main() {
	cfg := config.Load()

	dbpool, err := pgxpool.New(context.Background(), cfg.PostgresDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()

	paymentConn, err := grpc.Dial(cfg.PaymentGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer paymentConn.Close()

	paymentClient := &PaymentGRPCClient{
		client: pb.NewPaymentServiceClient(paymentConn),
	}

	repo := pgrepo.NewOrderRepository(dbpool)
	orderCache := cache.NewOrderCache(cfg.RedisAddr, cfg.OrderCacheTTL)
	uc := usecase.NewOrderUsecase(repo, paymentClient, orderCache)

	go startGRPCServer(cfg.OrderGRPCPort, uc)
	startHTTPServer(cfg.OrderHTTPPort, uc)
}

func startGRPCServer(port string, uc *usecase.OrderUsecase) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	pb.RegisterOrderServiceServer(server, grpcdelivery.NewOrderServer(uc))

	log.Println("order-service gRPC server on :" + port)
	log.Fatal(server.Serve(lis))
}

func startHTTPServer(port string, uc *usecase.OrderUsecase) {
	r := gin.Default()
	h := httpdelivery.NewHandler(uc)

	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/:id", h.GetOrder)
	r.PUT("/orders/:id/status", h.UpdateOrderStatus)

	log.Println("order-service HTTP server on :" + port)
	log.Fatal(r.Run(":" + port))
}
