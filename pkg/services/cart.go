package services

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/AthulKrishna2501/go-grpc-cart-service/pkg/db"
	"github.com/AthulKrishna2501/go-grpc-cart-service/pkg/models"
	"github.com/AthulKrishna2501/go-grpc-cart-service/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Server struct {
	H db.Handler
	pb.UnimplementedCartServiceServer
	ProductSvcClient pb.ProductServiceClient
}

func (s *Server) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.GetCartResponse, error) {
	var cartItems []models.CartItem
	userId := req.UserId

	if err := s.H.DB.Model(&models.CartItem{}).
		Select("cart_items.product_name, cart_items.quantity, cart_items.price, cart_items.product_id").
		Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("carts.user_id = ? AND cart_items.deleted_at IS NULL", userId).
		Find(&cartItems).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &pb.GetCartResponse{
				Message: "Cart is empty",
			}, status.Errorf(codes.NotFound, "Cart is empty or not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to fetch cart items: %v", err)
	}

	var items []*pb.CartItem
	for _, item := range cartItems {
		if item.ProductName == "" || item.Quantity == 0 {
			continue
		}

		items = append(items, &pb.CartItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    int64(item.Quantity),
			Price:       float64(item.Price),
		})
	}

	if len(items) == 0 {
		return &pb.GetCartResponse{
			Message: "Cart is empty",
			Items:   []*pb.CartItem{},
		}, nil
	}

	log.Print(items)

	return &pb.GetCartResponse{
		Message: "Cart retrieved successfully",
		Items:   items,
	}, nil
}

func (s *Server) AddToCart(ctx context.Context, req *pb.AddToCartRequest) (*pb.AddToCartResponse, error) {
	userId := req.UserId
	productId := req.ProductId

	res, err := s.ProductSvcClient.FindOne(ctx, &pb.FindOneRequest{Id: productId})
	if err != nil || res == nil || res.Data == nil {
		return &pb.AddToCartResponse{
			Message: "Product not found",
		}, status.Errorf(codes.NotFound, "Failed to fetch product: %v", err)
	}

	log.Print("Price", res.Data.Price)
	if res.Data.Stock < req.Quantity {
		return &pb.AddToCartResponse{
			Message: "Insufficient stock",
		}, nil
	}

	itemPrice := res.Data.Price
	log.Print("ItemsPrice:", itemPrice)

	var cart models.Cart
	if err := s.H.DB.Where("user_id = ?", userId).First(&cart).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		cart = models.Cart{
			UserID:     userId,
			ProductID:  uint64(productId),
			TotalPrice: 0,
		}
		if err := s.H.DB.Create(&cart).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to create cart: %v", err)
		}
	}

	var cartItem models.CartItem
	if err := s.H.DB.Where("cart_id = ? AND product_id = ?", cart.ID, productId).First(&cartItem).Error; err == nil {
		cartItem.Quantity += int(req.Quantity)
		cartItem.Price = itemPrice
		log.Print(cartItem.Price)
		s.H.DB.Save(&cartItem)

		cart.TotalPrice += float64(req.Quantity) * itemPrice
		s.H.DB.Save(&cart)

		return &pb.AddToCartResponse{Message: "Quantity updated"}, nil
	} else {
		cartItem = models.CartItem{
			CartID:      cart.ID,
			ProductName: res.Data.Name,
			ProductID:   productId,
			Quantity:    int(req.Quantity),
			Price:       itemPrice,
		}

		log.Print(cartItem)
		if err := s.H.DB.Create(&cartItem).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to add product to cart: %v", err)
		}

		cart.TotalPrice += float64(req.Quantity) * itemPrice
		s.H.DB.Save(&cart)
	}

	return &pb.AddToCartResponse{Message: "Product added to cart successfully"}, nil
}

func (s *Server) RemoveFromCart(ctx context.Context, req *pb.RemoveFromCartRequest) (*pb.RemoveFromCartResponse, error) {
	productId := req.ProductId

	var cartItem models.CartItem
	if err := s.H.DB.Where("product_id = ?", productId).First(&cartItem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Product not found in the cart")
		}
		return nil, status.Errorf(codes.Internal, "Failed to find cart item: %v", err)
	}

	if err := s.H.DB.Model(&cartItem).UpdateColumn("deleted_at", time.Now()).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to remove product from cart: %v", err)
	}

	var newTotalPrice float64
	if err := s.H.DB.Model(&models.CartItem{}).
		Where("cart_id = ? AND deleted_at IS NULL", cartItem.CartID).
		Select("COALESCE(SUM(price * quantity), 0)").Scan(&newTotalPrice).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update cart total: %v", err)
	}

	if err := s.H.DB.Model(&models.Cart{}).
		Where("id = ?", cartItem.CartID).
		Update("total_price", newTotalPrice).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update cart total price: %v", err)
	}

	return &pb.RemoveFromCartResponse{
		Message: "Product removed from cart successfully (soft delete)",
	}, nil
}

func (s *Server) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.ClearCartResponse, error) {
	if err := s.H.DB.Where("user_id = ?", req.UserId).Delete(&models.CartItem{}).Error; err != nil {
		return &pb.ClearCartResponse{
			Success: false,
			Message: "Failed to clear cart",
		}, err
	}

	return &pb.ClearCartResponse{
		Success: true,
		Message: "Cart cleared successfully",
	}, nil
}
