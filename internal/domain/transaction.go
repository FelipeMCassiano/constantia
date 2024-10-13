package domain

import "time"

type Transaction struct {
	ID          int       `json:"id"`
	IDSender    int       `json:"sender"`
	CPFSender   string    `json:"cpfsender"`
	CPFReciever int       `json:"cpfreciever"`
	Value       int       `json:"value"`
	CreatedAt   time.Time `json:"createdat"`
}
