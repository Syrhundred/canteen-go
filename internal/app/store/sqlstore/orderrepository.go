package sqlstore

import (
	"canteen-go/model"
	"time"
)

type OrderRepository struct {
	store *Store
}

func (o *OrderRepository) Create(order *model.Order) error {
	order.CreatedAt = time.Now()

	err := o.store.db.QueryRow(
		"INSERT INTO orders (user_id, createdAt, totalamount) VALUES ($1, $2, $3) RETURNING id",
		order.UserId,
		order.CreatedAt,
		order.TotalAmount,
	).Scan(&order.ID)
	if err != nil {
		return err
	}

	return nil
}

func (o *OrderRepository) Delete(id int) error {
	_, err := o.store.db.Exec("DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}
