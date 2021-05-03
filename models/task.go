package models

import (
	"database/sql"
	"errors"
	"log"
)

type Task struct {
	Id          string `json:"id"`
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"desc,omitempty" validate:"max=1000"`
	Status      bool   `json:"status"`
}

func (task *Task) Add(db *sql.DB) error {
	return db.QueryRow(
		"INSERT INTO tasks(name, description, status) VALUES($1, $2, $3) RETURNING id, name, description, status",
		task.Name,
		task.Description,
		task.Status).Scan(
		&task.Id,
		&task.Name,
		&task.Description,
		&task.Status)
}

func (task *Task) Get(db *sql.DB, id string) error {
	return db.QueryRow(
		"SELECT * FROM tasks WHERE id=$1",
		id).Scan(
		&task.Id,
		&task.Name,
		&task.Description,
		&task.Status)
}

func (task *Task) Delete(db *sql.DB, id string) error {
	result, err := db.Exec("DELETE FROM tasks WHERE id=$1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rows == 0 {
		return errors.New("not found")
	}

	return nil
}

func (task *Task) Update(db *sql.DB, id string) error {
	return db.QueryRow(
		"UPDATE tasks SET name=$2, description=$3, status=$4 WHERE id=$1 RETURNING id, name, description, status",
		id,
		task.Name,
		task.Description,
		task.Status).Scan(
		&task.Id,
		&task.Name,
		&task.Description,
		&task.Status)
}

func (task *Task) Complete(db *sql.DB, id string) error {
	return db.QueryRow(
		"UPDATE tasks SET status=true WHERE id=$1 RETURNING id, name, description, status",
		id).Scan(
		&task.Id,
		&task.Name,
		&task.Description,
		&task.Status)
}

func GetAll(db *sql.DB) ([]Task, error) {
	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tasks := []Task{}

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.Id, &task.Name, &task.Description, &task.Status); err != nil {
			return tasks, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
