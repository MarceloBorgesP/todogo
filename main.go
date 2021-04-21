package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/MarceloBorgesP/todogo/models"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	Todo   models.Todo
	DB     *sql.DB
}

func main() {
	a := App{}
	a.Initialize(
		os.Getenv("POSTGRES_HOSTNAME"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))
	a.Run(":8080")
}

func (a *App) Initialize(host, port, user, password, dbname string) {
	portNumber, _ := strconv.Atoi(port)
	connectionString :=
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, portNumber, user, password, dbname)

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
	rows, err := app.DB.Query("SELECT * FROM tasks")

	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err)
	}

	defer rows.Close()

	tasks := []models.Task{}

	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.Id, &task.Name, &task.Description, &task.Status); err != nil {
			respondWithJSON(w, http.StatusInternalServerError, err)
		}
		tasks = append(tasks, task)
	}

	respondWithJSON(w, http.StatusOK, &tasks)
}

func (app *App) createTask(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	newTask := models.Task{}

	err := decoder.Decode(&newTask)
	if err != nil {
		panic(err)
	}

	if errMessage := isValidInput(newTask); errMessage != nil {
		respondWithError(w, errMessage)

		return
	}

	err = app.DB.QueryRow("INSERT INTO tasks(name, description, status) VALUES($1, $2, $3) RETURNING id", newTask.Name, newTask.Description, newTask.Status).Scan(&newTask.Id)

	if err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusCreated, newTask)
	}
}

func (app *App) getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var task models.Task

	app.DB.QueryRow("SELECT * FROM tasks WHERE id=$1", vars["id"]).Scan(&task.Id, &task.Name, &task.Description, &task.Status)

	if task.Id == "" {
		respondWithJSON(w, http.StatusNotFound, nil)
	} else {
		respondWithJSON(w, http.StatusOK, task)
	}
}

func (app *App) updateTask(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var updatedTask models.Task
	err := decoder.Decode(&updatedTask)
	if err != nil {
		panic(err)
	}

	if errMessage := isValidInput(updatedTask); errMessage != nil {
		respondWithError(w, errMessage)

		return
	}

	vars := mux.Vars(r)

	_, err = app.DB.Exec("UPDATE tasks SET name=$2, description=$3, status=$4 WHERE id=$1", vars["id"], updatedTask.Name, updatedTask.Description, updatedTask.Status)

	if err == nil {
		respondWithJSON(w, http.StatusOK, updatedTask)
	} else {
		respondWithJSON(w, http.StatusNotFound, nil)
	}
}

func (app *App) deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, err := app.DB.Exec("DELETE FROM tasks WHERE id=$1", vars["id"])

	if err == nil {
		respondWithJSON(w, http.StatusNoContent, nil)
	} else {
		respondWithJSON(w, http.StatusNotFound, nil)
	}

}

func (app *App) completeTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, err := app.DB.Exec("UPDATE tasks SET status=true WHERE id=$1", vars["id"])

	if err == nil {
		respondWithJSON(w, http.StatusNoContent, nil)
	} else {
		respondWithJSON(w, http.StatusNotFound, nil)
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

func isValidInput(task models.Task) map[string]string {
	validate := validator.New()
	err := validate.Struct(task)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		return map[string]string{"error": validationErrors.Error()}
	}

	return nil
}
