package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

var (
	tasks    []Task
	fileName = "tasks.json"
	mu       sync.Mutex
)

func loadTasks() {
	data, err := os.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			tasks = []Task{}
			return
		}
		fmt.Println("Ошибка чтения:", err)
		os.Exit(1)
	}
	json.Unmarshal(data, &tasks)
}

func saveTasks() {
	data, _ := json.MarshalIndent(tasks, "", "  ")
	os.WriteFile(fileName, data, 0644)
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func addTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var req struct {
		Name string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	id := 1
	if len(tasks) > 0 {
		id = tasks[len(tasks)-1].ID + 1
	}
	tasks = append(tasks, Task{ID: id, Name: req.Name, Done: false})
	saveTasks()
	w.WriteHeader(http.StatusCreated)
}

func toggleTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var req struct {
		ID int `json:"id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	for i := range tasks {
		if tasks[i].ID == req.ID {
			tasks[i].Done = !tasks[i].Done
			break
		}
	}
	saveTasks()
	w.WriteHeader(http.StatusOK)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	for i, t := range tasks {
		if t.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			break
		}
	}
	saveTasks()
	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/api/tasks", getTasks)
	http.HandleFunc("/api/add", addTask)
	http.HandleFunc("/api/toggle", toggleTask)
	http.HandleFunc("/api/delete", deleteTask)

	// отдаём фронтенд из папки frontend/
	http.Handle("/", http.FileServer(http.Dir("frontend")))

	fmt.Println("🚀 Сервер запущен на http://localhost:8081")
	http.ListenAndServe(":8081", nil)
}
