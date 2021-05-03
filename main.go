package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/MarceloBorgesP/todogo/models"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func main() {
	a := App{}
	a.Initialize(
		"localhost",
		"5432",
		"postgres",
		"postgres",
		"postgres")
	a.Run(":8080")
}

func (a *App) Initialize(host, port, user, password, dbname string) {
	portNumber, _ := strconv.Atoi(port)
	connectionString :=
		fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host,
			portNumber,
			user,
			password,
			dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

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
	a.Router.HandleFunc("/task/{id}/complete", a.completeTask).Methods("POST")
}

func (app *App) getTasks(w http.ResponseWriter, r *http.Request) {
	if tasks, err := models.GetAll(app.DB); err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err)
	} else {
		respondWithJSON(w, http.StatusOK, tasks)
	}
}

func (app *App) getTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	vars := mux.Vars(r)

	if err := task.Get(app.DB, vars["id"]); err != nil {
		respondWithJSON(w, http.StatusNotFound, nil)
	} else {
		respondWithJSON(w, http.StatusOK, task)
	}
}

func (app *App) createTask(w http.ResponseWriter, r *http.Request) {
	newTask, err := DecodeTask(r.Body)
	if err != nil {
		respondWithError(w, err)

		return
	}

	if err := newTask.Add(app.DB); err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusCreated, newTask)
	}
}

func (app *App) updateTask(w http.ResponseWriter, r *http.Request) {
	updatedTask, err := DecodeTask(r.Body)
	if err != nil {
		respondWithError(w, err)

		return
	}

	vars := mux.Vars(r)

	if errMessage := updatedTask.Update(app.DB, vars["id"]); errMessage != nil {
		respondWithJSON(w, http.StatusNotFound, nil)
	} else {
		respondWithJSON(w, http.StatusOK, updatedTask)
	}
}

func (app *App) deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var task models.Task

	if err := task.Delete(app.DB, vars["id"]); err != nil {
		respondWithJSON(w, http.StatusNotFound, nil)
	} else {
		respondWithJSON(w, http.StatusNoContent, nil)
	}

}

func (app *App) completeTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var task models.Task

	if err := task.Complete(app.DB, vars["id"]); err != nil {
		respondWithJSON(w, http.StatusNotFound, nil)
	} else {
		respondWithJSON(w, http.StatusNoContent, nil)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, errMessage interface{}) {
	w.Header().Add("Content-Type", "application/json")
	respondWithJSON(w, http.StatusBadRequest, errMessage)
}

func DecodeTask(body io.ReadCloser) (models.Task, map[string]string) {
	decoder := json.NewDecoder(body)
	task := models.Task{}

	err := decoder.Decode(&task)
	if err != nil {
		panic(err)
	}

	errMessage := isValidInput(task)
	return task, errMessage
}

func isValidInput(task models.Task) map[string]string {
	validate := validator.New()
	err := validate.Struct(task)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		return map[string]string{"error": validationErrors.Error()}
	}

	return nil
}
