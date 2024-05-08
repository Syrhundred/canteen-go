package store

type Store interface {
	User() UserRepository
	MenuItem() MenuItemRepository
	OrderRepository() OrderRepository
	OrderItemRepository() OrderItemRepository
}
