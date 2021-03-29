package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/lithammer/shortuuid/v3"
)

type Task struct {
	Id          string `json:"id"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"desc" validate:"required"`
	Status      string `json:"status" validate:"required"`
}

type TodoHandler struct {
	mu    sync.Mutex
	Tasks []Task
}

func main() {
	http.Handle("/todo", new(TodoHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (todo *TodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	todo.mu.Lock()
	defer todo.mu.Unlock()

	switch r.Method {
	case "GET":
		todo.getTaskHandler(w, r)
		break
	case "POST":
		todo.postTaskHandler(w, r)
		break
	default:
		json.NewEncoder(w).Encode(todo.Tasks)
		break
	}

}

func (todo *TodoHandler) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(todo.Tasks)
}

func (todo *TodoHandler) postTaskHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var newTask Task
	err := decoder.Decode(&newTask)
	if err != nil {
		panic(err)
	}
	newTask.Id = shortuuid.New()

	// err2 := validate.Validate(newTask)
	// if err2 != nil {
	// 	json.NewEncoder(err2).Encode(todo.Tasks)
	// }

	todo.Tasks = append(todo.Tasks, newTask)
	json.NewEncoder(w).Encode(newTask)
}
