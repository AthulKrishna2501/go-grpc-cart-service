package client

import (
	"fmt"

	"github.com/AthulKrishna2501/go-grpc-cart-service/pkg/config"
	"github.com/AthulKrishna2501/go-grpc-cart-service/pkg/pb"
	"google.golang.org/grpc"
)

type ServiceClient struct {
	Client pb.ProductServiceClient
}

func InitServiceClient(c *config.Config) pb.ProductServiceClient {
	cc, err := grpc.Dial(c.ProductSvc, grpc.WithInsecure())

	if err != nil {
		fmt.Println("Could not connect:", err)
	}

	return pb.NewProductServiceClient(cc)
}
