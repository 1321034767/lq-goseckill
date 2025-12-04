package models

type Message struct {
	ProductID int64
	UserID    int64
}

func NewMessage(productID, userID int64) *Message {
	return &Message{ProductID: productID, UserID: userID}
}
