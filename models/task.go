package models

import "sync"

type Task struct {
	Id          string `json:"id"`
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"desc,omitempty" validate:"max=1000"`
	Status      bool   `json:"status"`
}

type Todo struct {
	mu    sync.Mutex
	Tasks []Task
}
