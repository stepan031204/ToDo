package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

var tasks []Task
var fileName = "task.json"

func loadTasks() {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			tasks = []Task{}
			return
		}
		fmt.Println("fail", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &tasks)
}

func saveTasks() {
	data, _ := json.MarshalIndent(tasks, "", "  ")
	ioutil.WriteFile(fileName, data, 0644)
}
func listTasks() {
	fmt.Println("Список задач:")
	for _, t := range tasks {
		status := "❌"
		if t.Done {
			status = "✅"
		}
		fmt.Printf("%d. %s [%s]\n", t.ID, t.Name, status)
	}
}

func addTask(name string) {
	id := 1
	if len(tasks) > 0 {
		id = tasks[len(tasks)-1].ID + 1
	}
	tasks = append(tasks, Task{ID: id, Name: name, Done: false})
	saveTasks()
}

func main() {
	loadTasks()

	for {
		fmt.Println("1)Список задач \n2)Добавить \n3)Выйти")
		var command string
		fmt.Scan(&command)

		switch command {
		case "1":
			listTasks()
		case "2":
			fmt.Println("Введите название задачи:")
			var name string
			fmt.Scanln(&name)
			addTask(name)
			fmt.Println("Задача добавлена!")
		case "3":
			return
		default:
			fmt.Println("Неизвестная команда")
		}
	}
}
