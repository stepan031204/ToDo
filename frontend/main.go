package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Сервер запущен на порту 8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Создаём контекст с таймаутом в 5 секунд
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel() // Важно! Всегда вызываем cancel

	// Симулируем долгую операцию
	result := make(chan string, 1)
	go func() {
		time.Sleep(3 * time.Second) // Работаем 3 секунды
		result <- "Операция завершена"
	}()

	select {
	case <-ctx.Done():
		http.Error(w, "Таймаут запроса", http.StatusRequestTimeout)
		return
	case res := <-result:
		fmt.Fprintf(w, res)
	}
}
