package store

type Store interface {
	User() UserRepository
	MenuItem() MenuItemRepository
	Order() OrderRepository
	OrderItem() OrderItemRepository
}
