package handler

import (
	recordtransport "github.com/al-tokarev/gophkeeper/internal/modules/record/transport"
	usertransport "github.com/al-tokarev/gophkeeper/internal/modules/user/transport"
)

// Handler собирает HTTP хендлеры для всех модулей
// Не содержит логики, используется как точка сбора
type Handler struct {
	User   *usertransport.UserHandler
	Record *recordtransport.RecordHandler
}

// NewHandler создаёт агрегатор хендлеров для передачи в роутер
func NewHandler(
	u *usertransport.UserHandler,
	r *recordtransport.RecordHandler,
) *Handler {
	return &Handler{
		User:   u,
		Record: r,
	}
}
