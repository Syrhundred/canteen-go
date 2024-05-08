package store

import "canteen-go/model"

type UserRepository interface {
	Create(*model.User) error
	Find(int) (*model.User, error)
	FindByEmail(string) (*model.User, error)
}

type MenuItemRepository interface {
	Create(*model.MenuItem) error
	Delete(id int) error
	GetPrice(id int) int
}

type OrderRepository interface {
	Create(*model.Order) error
	Delete(id int) error
}

type OrderItemRepository interface {
	Create(*model.OrderItem) error
}
