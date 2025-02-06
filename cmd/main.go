package main

import (
	"fmt"
	"log"
	"net"

	"github.com/AthulKrishna2501/go-grpc-cart-service/pkg/config"
	"github.com/AthulKrishna2501/go-grpc-cart-service/pkg/db"
	pb "github.com/AthulKrishna2501/go-grpc-cart-service/pkg/pb"

	"github.com/AthulKrishna2501/go-grpc-cart-service/pkg/services"

	"google.golang.org/grpc"
)

func main() {
	c, err := config.LoadConfig()

	if err != nil {
		log.Fatalln("Failed at config", err)
	}

	h := db.Init(c.DBUrl)

	lis, err := net.Listen("tcp", c.Port)

	if err != nil {
		log.Fatalln("Failed to listing:", err)
	}

	fmt.Println("Cart Svc on", c.Port)

	productSvc, err := grpc.Dial(c.ProductSvc, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}

	defer productSvc.Close()

	productServiceClient := pb.NewProductServiceClient(productSvc)

	s := services.Server{
		H:                h,
		ProductSvcClient: productServiceClient,
	}

	grpcServer := grpc.NewServer()

	pb.RegisterCartServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}
