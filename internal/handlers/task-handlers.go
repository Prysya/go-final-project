package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Prysya/go-final-project/pkg/api"
	"github.com/Prysya/go-final-project/pkg/constants"
	"github.com/Prysya/go-final-project/pkg/models"
	"github.com/Prysya/go-final-project/pkg/repository"
)

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	repo, err := repository.NewTaskRepository()
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetTasks(w, r, repo)
	case http.MethodPost:
		handleCreateTask(w, r, repo)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleGetTasks(w http.ResponseWriter, r *http.Request, repo *repository.TaskRepository) {
	date := r.URL.Query().Get("date")

	var tasks []models.Task
	var err error

	if date != "" {
		tasks, err = repo.GetByDate(date)
	} else {
		tasks, err = repo.GetAll()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func handleCreateTask(w http.ResponseWriter, r *http.Request, repo *repository.TaskRepository) {
	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSONError(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeJSONError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	now := time.Now()
	today := now.Format(constants.DateFormat)

	if task.Date == "" {
		task.Date = today
	} else {
		_, err := time.Parse(constants.DateFormat, task.Date)
		if err != nil {
			writeJSONError(w, fmt.Sprintf("Дата представлена в формате, отличном от %d", constants.DateFormat), http.StatusBadRequest)
			return
		}
	}

	taskTime, _ := time.Parse(constants.DateFormat, task.Date)

	if taskTime.Before(now.Truncate(24 * time.Hour)) {
		if task.Repeat == "" {
			task.Date = today
		} else {
			nextDate, err := api.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				writeJSONError(w, "Правило повторения указано в неправильном формате", http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	} else {
		if task.Repeat != "" {
			_, err := api.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				writeJSONError(w, "Правило повторения указано в неправильном формате", http.StatusBadRequest)
				return
			}
		}
	}

	id, err := repo.Create(&task)
	if err != nil {
		writeJSONError(w, "Ошибка при создании задачи: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id": strconv.FormatInt(id, 10),
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]string{
		"error": message,
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
