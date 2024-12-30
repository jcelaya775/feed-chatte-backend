package models

import "time"

type Event struct {
	Id      string    `json:"id"`
	UserId  string    `json:"userId"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}
