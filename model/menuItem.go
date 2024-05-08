package model

type MenuItem struct {
	ID          int    `json:"id"`          // Уникальный идентификатор
	Name        string `json:"name"`        // Название блюда
	Price       int    `json:"price"`       // Цена
	Description string `json:"description"` // Описание блюда
}
