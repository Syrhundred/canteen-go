package sqlstore

import "canteen-go/model"

type OrderItemRepository struct {
	store *Store
}

func (i *OrderItemRepository) Create(item *model.OrderItem) error {
	return i.store.db.QueryRow(
		"INSERT INTO orderitem (order_id, menu_item_id, quantity) VALUES ($1, $2, $3) RETURNING id",
		item.OrderId,
		item.MenuItemId,
		item.Quantity).Scan(&item.ID)
}
