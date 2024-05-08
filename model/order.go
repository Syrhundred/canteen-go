package model

import "time"

type Order struct {
	ID          int       `json:"id"`
	UserId      int       `json:"user_id"`
	CreatedAt   time.Time `json:"-"`
	TotalAmount int       `json:"-"`
}
