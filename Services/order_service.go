package services

import (
	"github.com/mikemeh/ecommerce-api/internal/models"
	"github.com/mikemeh/ecommerce-api/pkg/errors"
	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) CreateOrder(order *models.Order) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return errors.Wrap(err, "failed to create order")
		}

		for _, item := range order.OrderItems {
			var product models.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return errors.Wrap(err, "failed to get product")
			}

			if product.Stock < item.Quantity {
				return errors.BadRequest("insufficient stock for product: %d", item.ProductID)
			}

			product.Stock -= item.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return errors.Wrap(err, "failed to update product stock")
			}
		}

		return nil
	})
}

func (s *OrderService) GetOrdersByUserID(userID uint) ([]models.Order, error) {
	var orders []models.Order
	if err := s.db.Where("user_id = ?", userID).Preload("OrderItems").Find(&orders).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get orders")
	}
	return orders, nil
}

func (s *OrderService) GetOrderByID(id uint) (*models.Order, error) {
	var order models.Order
	if err := s.db.Preload("OrderItems").First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("order not found")
		}
		return nil, errors.Wrap(err, "failed to get order")
	}
	return &order, nil
}

func (s *OrderService) UpdateOrderStatus(id uint, status string) error {
	if err := s.db.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return errors.Wrap(err, "failed to update order status")
	}
	return nil
}

func (s *OrderService) CancelOrder(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Preload("OrderItems").First(&order, id).Error; err != nil {
			return errors.Wrap(err, "failed to get order")
		}

		if order.Status != "Pending" {
			return errors.BadRequest("only pending orders can be cancelled")
		}

		for _, item := range order.OrderItems {
			var product models.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return errors.Wrap(err, "failed to get product")
			}

			product.Stock += item.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return errors.Wrap(err, "failed to update product stock")
			}
		}

		if err := tx.Model(&order).Update("status", "Cancelled").Error; err != nil {
			return errors.Wrap(err, "failed to update order status")
		}

		return nil
	})
}
