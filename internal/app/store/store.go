package store

type Store interface {
	User() UserRepository
	MenuItem() MenuItemRepository
}
