package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

var tasks []Task
var fileName = "task.json"

func loadTasks() {
	data, err := os.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			tasks = []Task{}
			return
		}
		fmt.Println("Ошибка чтения файла:", err)
		return
	}
	json.Unmarshal(data, &tasks)
}

func saveTasks() {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		fmt.Println("Ошибка при сохранении:", err)
		return
	}
	os.WriteFile(fileName, data, 0644)
}

func main() {
	loadTasks()

	a := app.New()
	w := a.NewWindow("To-Do List")
	w.Resize(fyne.NewSize(400, 500))

	// Список задач
	list := widget.NewList(
		func() int { return len(tasks) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			status := "[ ]"
			if tasks[i].Done {
				status = "[x]"
			}
			o.(*widget.Label).SetText(fmt.Sprintf("%d. %s %s", tasks[i].ID, tasks[i].Name, status))
		},
	)

	// Поле ввода для новой задачи
	input := widget.NewEntry()
	input.SetPlaceHolder("Введите название задачи")

	// Кнопка добавить
	addBtn := widget.NewButton("Добавить", func() {
		name := strings.TrimSpace(input.Text)
		if name == "" {
			return
		}
		id := 1
		if len(tasks) > 0 {
			id = tasks[len(tasks)-1].ID + 1
		}
		tasks = append(tasks, Task{ID: id, Name: name, Done: false})
		saveTasks()
		list.Refresh()
		input.SetText("")
	})

	// Кнопка удалить
	delBtn := widget.NewButton("Удалить по ID", func() {
		dialog.ShowEntryDialog("Удалить задачу", "Введите ID задачи", func(idStr string) {
			var id int
			fmt.Sscanf(idStr, "%d", &id)
			index := -1
			for i, t := range tasks {
				if t.ID == id {
					index = i
					break
				}
			}
			if index != -1 {
				tasks = append(tasks[:index], tasks[index+1:]...)
				// пересчет ID
				for i := range tasks {
					tasks[i].ID = i + 1
				}
				saveTasks()
				list.Refresh()
			}
		}, w)
	})

	// Кнопка изменить название
	editBtn := widget.NewButton("Переименовать", func() {
		dialog.ShowEntryDialog("Переименовать задачу", "Введите ID и новое название через ':'", func(text string) {
			parts := strings.SplitN(text, ":", 2)
			if len(parts) != 2 {
				return
			}
			var id int
			fmt.Sscanf(parts[0], "%d", &id)
			newName := strings.TrimSpace(parts[1])
			for i := range tasks {
				if tasks[i].ID == id {
					tasks[i].Name = newName
					saveTasks()
					list.Refresh()
					break
				}
			}
		}, w)
	})

	// Кнопка изменить статус
	statusBtn := widget.NewButton("Сменить статус", func() {
		dialog.ShowEntryDialog("Изменить статус", "Введите ID задачи", func(idStr string) {
			var id int
			fmt.Sscanf(idStr, "%d", &id)
			for i := range tasks {
				if tasks[i].ID == id {
					tasks[i].Done = !tasks[i].Done
					saveTasks()
					list.Refresh()
					break
				}
			}
		}, w)
	})

	// Сборка интерфейса
	w.SetContent(container.NewVBox(
		list,
		input,
		container.NewHBox(addBtn, delBtn),
		container.NewHBox(editBtn, statusBtn),
	))

	w.ShowAndRun()
}
