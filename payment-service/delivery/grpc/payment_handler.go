package grpc

import (
	"context"

	pb "github.com/Yessenchik/generated/orderpayment"
	"github.com/Yessenchik/payment-service/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUsecase
}

func NewPaymentServer(uc *usecase.PaymentUsecase) *PaymentServer {
	return &PaymentServer{uc: uc}
}

func (s *PaymentServer) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	ok, msg, err := s.uc.ProcessPayment(
		ctx,
		req.OrderId,
		req.Amount,
		req.CustomerEmail,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.PaymentResponse{
		Success: ok,
		Message: msg,
	}, nil
}

func (s *PaymentServer) ListPayments(ctx context.Context, req *pb.ListPaymentsRequest) (*pb.ListPaymentsResponse, error) {
	payments, err := s.uc.ListPayments(ctx, req.Status)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.ListPaymentsResponse{
		Payments: make([]*pb.PaymentItem, 0, len(payments)),
	}

	for _, p := range payments {
		resp.Payments = append(resp.Payments, &pb.PaymentItem{
			Id:        p.ID,
			OrderId:   p.OrderID,
			Amount:    p.Amount,
			Status:    p.Status,
			CreatedAt: timestamppb.New(p.CreatedAt),
		})
	}

	return resp, nil
}
