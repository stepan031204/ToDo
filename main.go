package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
)

type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

var (
	tasks    []Task
	mu       sync.Mutex
	dataFile = "tasks.json"
)

func loadTasks() {
	data, err := os.ReadFile(dataFile)
	if err == nil {
		json.Unmarshal(data, &tasks)
	}
}

func saveTasks() {
	data, _ := json.MarshalIndent(tasks, "", "  ")
	os.WriteFile(dataFile, data, 0644)
}

func listTasks(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func addTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var t Task
	json.NewDecoder(r.Body).Decode(&t)
	t.ID = len(tasks) + 1
	tasks = append(tasks, t)
	saveTasks()
	json.NewEncoder(w).Encode(t)
}

func toggleTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var req struct{ ID int }
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

func main() {
	loadTasks()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/list", listTasks)
	http.HandleFunc("/api/add", addTask)
	http.HandleFunc("/api/toggle", toggleTask)

	http.ListenAndServe(":8081", nil)
}
