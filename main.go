package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

var (
	db *sql.DB
	mu sync.Mutex
)

func initLogger() {
	f, err := os.OpenFile("todo.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Не удалось открыть лог-файл: %v", err)
	}
	log.SetOutput(f)
	log.SetPrefix("[ToDo]")
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "tasks.db")
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			done INTEGER NOT NULL
		);`)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы: %v", err)
	}
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	mu.Lock()
	defer mu.Unlock()

	rows, err := db.QueryContext(ctx, "SELECT id, name, done FROM tasks")
	if err != nil {
		http.Error(w, "Ошибка запроса", http.StatusInternalServerError)
		log.Printf("Ошибка запроса: %v", err)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var done int
		if err := rows.Scan(&t.ID, &t.Name, &done); err != nil {
			http.Error(w, "Ошибка чтения данных", http.StatusInternalServerError)
			log.Printf("Ошибка чтения данных: %v", err)
			return
		}
		t.Done = done != 0
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, "Ошибка отправки JSON", http.StatusInternalServerError)
		log.Printf("Ошибка отправки JSON: %v", err)
		return
	}

}

func addTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	mu.Lock()
	defer mu.Unlock()

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "Неверный ввод", http.StatusBadRequest)
		log.Printf("Невернный ввод пользователя: %v", err)
		return
	}
	_, err := db.ExecContext(ctx, "INSERT INTO tasks(name, done) VALUES(?, ?)", req.Name, 0)
	if err != nil {
		http.Error(w, "Ошибка вставки", http.StatusInternalServerError)
		log.Printf("Ошибка вставки в базу данных: %v", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func toggleTask(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	mu.Lock()
	defer mu.Unlock()

	var req struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный ввод", http.StatusBadRequest)
		log.Printf("Невернный ввод пользователя: %v", err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Ошибка транзакции", http.StatusInternalServerError)
		log.Panic("Ошибка транзакции: ", err)
		return
	}
	defer tx.Rollback()

	var done int
	err = tx.QueryRowContext(ctx, "SELECT done FROM tasks WHERE id = ?", req.ID).Scan(&done)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		log.Printf("Задача не найдена: %v", err)
		return
	}

	_, err = tx.ExecContext(ctx, "UPDATE tasks SET done = ? WHERE id = ?", 1-done, req.ID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		log.Printf("Ошибка обновления базы данных: %v", err)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Ошибка транзации", http.StatusInternalServerError)
		log.Printf("Ошибка транзации: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	mu.Lock()
	defer mu.Unlock()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	_, err = db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	initLogger()
	initDB()
	defer db.Close()
	http.HandleFunc("/api/tasks", getTasks)
	http.HandleFunc("/api/add", addTask)
	http.HandleFunc("/api/toggle", toggleTask)
	http.HandleFunc("/api/delete", deleteTask)

	http.Handle("/", http.FileServer(http.Dir("frontend")))

	fmt.Println("🚀 Сервер запущен на http://localhost:8081")
	http.ListenAndServe(":8081", nil)
}
