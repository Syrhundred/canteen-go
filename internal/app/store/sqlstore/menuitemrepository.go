package sqlstore

import "canteen-go/model"

type MenuItemRepository struct {
	store *Store
}

func (r *MenuItemRepository) Create(m *model.MenuItem) error {
	query := `INSERT INTO menu_items (name, price, description) VALUES ($1, $2, $3) RETURNING id`
	if err := r.store.db.QueryRow(query, m.Name, m.Price, m.Description).Scan(&m.ID); err != nil {
		return err
	}
	return nil
}

func (r *MenuItemRepository) Delete(id int) error {
	_, err := r.store.db.Exec("DELETE FROM menu_items WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}
