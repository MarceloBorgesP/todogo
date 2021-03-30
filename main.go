package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"

	"github.com/lithammer/shortuuid/v3"
)

type App struct {
	Router *mux.Router
	Todo   Todo
}

type Todo struct {
	mu    sync.Mutex
	Tasks []Task
}

type Task struct {
	Id          string `json:"id"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"desc" validate:"required"`
	Status      string `json:"status" validate:"required"`
}

func main() {
	a := App{}
	a.Initialize()
	a.Run(":8080")
}

func (a *App) Initialize() {
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/task", a.getTasks).Methods("GET")
	a.Router.HandleFunc("/task", a.createTask).Methods("POST")
	a.Router.HandleFunc("/task/{id}", a.getTask).Methods("GET")
	a.Router.HandleFunc("/task/{id}", a.updateTask).Methods("PUT")
	a.Router.HandleFunc("/task/{id}", a.deleteTask).Methods("DELETE")
}

func (app *App) getTasks(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, app.Todo.Tasks)
}

func (app *App) createTask(w http.ResponseWriter, r *http.Request) {
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

	app.Todo.Tasks = append(app.Todo.Tasks, newTask)
	respondWithJSON(w, http.StatusCreated, newTask)
}

func (app *App) getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	for _, task := range app.Todo.Tasks {
		if task.Id == vars["id"] {
			respondWithJSON(w, http.StatusOK, task)
			return
		}
	}

	respondWithJSON(w, http.StatusNotFound, nil)
}

func (app *App) updateTask(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var updatedTask Task
	err := decoder.Decode(&updatedTask)
	if err != nil {
		panic(err)
	}

	vars := mux.Vars(r)
	for i, task := range app.Todo.Tasks {
		if task.Id == vars["id"] {
			app.Todo.Tasks[i].Name = updatedTask.Name
			app.Todo.Tasks[i].Description = updatedTask.Description
			app.Todo.Tasks[i].Status = updatedTask.Status
			respondWithJSON(w, http.StatusOK, app.Todo.Tasks[i])

			return
		}
	}

	respondWithJSON(w, http.StatusNotFound, nil)
}

func (app *App) deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	for i, task := range app.Todo.Tasks {
		if task.Id == vars["id"] {
			app.Todo.Tasks[i] = app.Todo.Tasks[len(app.Todo.Tasks)-1]
			app.Todo.Tasks = app.Todo.Tasks[:len(app.Todo.Tasks)-1]
			respondWithJSON(w, http.StatusNoContent, nil)
			return
		}
	}

	respondWithJSON(w, http.StatusNotFound, nil)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
