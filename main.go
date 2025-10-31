package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
)

type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

var tasks []Task
var fileName = "tasks.json"

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

func listHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(htmlPage))
	tmpl.Execute(w, tasks)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		if name == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		id := 1
		if len(tasks) > 0 {
			id = tasks[len(tasks)-1].ID + 1
		}
		tasks = append(tasks, Task{ID: id, Name: name})
		saveTasks()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func doneHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Done = !tasks[i].Done
			saveTasks()
			break
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	index := -1
	for i, t := range tasks {
		if t.ID == id {
			index = i
			break
		}
	}
	if index != -1 {
		tasks = append(tasks[:index], tasks[index+1:]...)
		saveTasks()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

const htmlPage = `
<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<title>To-Do List</title>
<style>
body {
	font-family: Arial, sans-serif;
	max-width: 600px;
	margin: 40px auto;
	background: #f5f5f5;
	padding: 20px;
	border-radius: 12px;
	box-shadow: 0 0 10px rgba(0,0,0,0.1);
}
h1 { text-align: center; }
form { display: flex; margin-bottom: 20px; }
input[type=text] {
	flex: 1; padding: 10px;
	border: 1px solid #ccc;
	border-radius: 6px;
}
button {
	margin-left: 10px;
	padding: 10px 15px;
	background: #007BFF;
	color: white;
	border: none;
	border-radius: 6px;
	cursor: pointer;
}
button:hover { background: #0056b3; }
.task {
	display: flex;
	justify-content: space-between;
	padding: 8px;
	background: white;
	border-radius: 6px;
	margin-bottom: 8px;
}
.done { text-decoration: line-through; color: gray; }
a {
	text-decoration: none;
	color: #007BFF;
	margin-left: 10px;
}
a:hover { text-decoration: underline; }
</style>
</head>
<body>
<h1>📝 Мои задачи</h1>

<form method="POST" action="/add">
	<input type="text" name="name" placeholder="Новая задача..." required>
	<button type="submit">Добавить</button>
</form>

{{if .}}
	{{range .}}
	<div class="task">
		<span class="{{if .Done}}done{{end}}">{{.ID}}. {{.Name}}</span>
		<div>
			<a href="/done?id={{.ID}}">✔</a>
			<a href="/delete?id={{.ID}}">🗑</a>
		</div>
	</div>
	{{end}}
{{else}}
	<p>Пока нет задач ✨</p>
{{end}}
</body>
</html>
`

func main() {
	loadTasks()
	http.HandleFunc("/", listHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/done", doneHandler)
	http.HandleFunc("/delete", deleteHandler)

	fmt.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8081", nil)
}
