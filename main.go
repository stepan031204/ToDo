package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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

func saveTasks(tasks []Task) {
	data, err := json.MarshalIndent(tasks, "", "  ")

	if err != nil {
		fmt.Println("Ошибка при сохранение задачи", err)
	}
	ioutil.WriteFile(fileName, data, 0644)
}

func listTasks() {
	if len(tasks) == 0 {
		fmt.Println("Список пуст")
	} else {
		fmt.Println("Список задач:")
	}
	for _, t := range tasks {
		status := "[ ]"
		if t.Done {
			status = "[x]"
		}
		fmt.Printf("%d. %s %s\n", t.ID, t.Name, status)
	}

}

func addTask() []Task {
	id := 0

	fmt.Println("Введите название задачи:")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if len(tasks) > 0 {
		id = tasks[len(tasks)-1].ID + 1
	}
	tasks = append(tasks, Task{ID: id, Name: name, Done: false})
	saveTasks(tasks)
	return tasks
}

func removeTask(tasks []Task) []Task {
	index := -1
	var id int

	fmt.Println("Введите id")
	fmt.Scanln(&id)
	for i, t := range tasks {
		if t.ID == id {
			index = i
			break
		}
	}
	if index == -1 {
		fmt.Println("Задача с таким ID не найдена.")
		return tasks
	}

	tasks = append(tasks[:index], tasks[index+1:]...)
	for i := range tasks {
		tasks[i].ID = i + 1
	}
	saveTasks(tasks)
	return tasks
}
func changeStatus(tasks []Task) []Task {
	var id int
	fmt.Println("Введите id")
	fmt.Scanln(&id)

	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Done = true
		}
	}
	saveTasks(tasks)
	return tasks
}

func renameTask(tasks []Task) []Task {
	var id int
	fmt.Println("Введите id")
	fmt.Scanln(&id)

	for i := range tasks {
		if tasks[i].ID == id {
			fmt.Println("Введите новое название")
			reader := bufio.NewReader(os.Stdin)
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			tasks[i].Name = name

		}
	}
	saveTasks(tasks)
	return tasks
}

func main() {

	loadTasks()

	for {
		var command string

		fmt.Println("\n1)Список \n2)Добавить \n3)Удалить \n4)Изменить \n5)Статус \n6)Загрузить \n7)Выйти")
		fmt.Scan(&command)
		fmt.Println("")

		switch command {
		case "1":
			listTasks()
		case "2":
			addTask()
		case "3":
			tasks = removeTask(tasks)
		case "4":
			tasks = renameTask(tasks)
		case "5":
			tasks = changeStatus(tasks)
		case "6":
			loadTasks()
		case "7":
			return
		default:
			fmt.Println("Неизвестная команда")
		}
	}
}
