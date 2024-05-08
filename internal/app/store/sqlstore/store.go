package sqlstore

import (
	"canteen-go/internal/app/store"
	"database/sql"
	_ "github.com/lib/pq"
)

type Store struct {
	db                 *sql.DB
	UserRepository     *UserRepository
	MenuItemRepository *MenuItemRepository
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) User() store.UserRepository {
	if s.UserRepository != nil {
		return s.UserRepository
	}

	s.UserRepository = &UserRepository{
		store: s,
	}

	return s.UserRepository
}

func (s *Store) MenuItem() store.MenuItemRepository {
	if s.MenuItemRepository != nil {
		return s.MenuItemRepository
	}

	s.MenuItemRepository = &MenuItemRepository{
		store: s,
	}

	return s.MenuItemRepository
}
