package grpc

import (
	"time"

	pb "github.com/Yessenchik/generated/orderpayment"
	"github.com/Yessenchik/order-service/usecase"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	uc *usecase.OrderUsecase
}

func NewOrderServer(uc *usecase.OrderUsecase) *OrderServer {
	return &OrderServer{uc: uc}
}

func (s *OrderServer) SubscribeToOrderUpdates(
	req *pb.OrderRequest,
	stream pb.OrderService_SubscribeToOrderUpdatesServer,
) error {
	ctx := stream.Context()

	order, err := s.uc.GetOrder(ctx, req.OrderId)
	if err != nil {
		return err
	}

	lastUpdated := order.UpdatedAt

	if err := stream.Send(&pb.OrderStatusUpdate{
		//OrderId:   changedOrder.ID,
		Status:    order.Status,
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			changedOrder, err := s.uc.GetOrder(ctx, req.OrderId)
			if err != nil {
				return err
			}

			if changedOrder.UpdatedAt.After(lastUpdated) {
				lastUpdated = changedOrder.UpdatedAt

				if err := stream.Send(&pb.OrderStatusUpdate{
					//OrderId:   changedOrder.ID,
					Status:    changedOrder.Status,
					UpdatedAt: timestamppb.New(changedOrder.UpdatedAt),
				}); err != nil {
					return err
				}
			}
		}
	}
}
