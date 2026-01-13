package handlers

import (
	"net/http"
	"strconv"

	"github.com/Prysya/go-final-project/pkg/api"
	"github.com/Prysya/go-final-project/pkg/models"
	"github.com/Prysya/go-final-project/pkg/repository"
)

type TasksResp struct {
	Tasks []*models.Task `json:"tasks"`
}

const DefaultTaskLimit = 50

func TasksHandler(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetTasksCollection(w, r, repo)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func handleGetTasksCollection(w http.ResponseWriter, r *http.Request, repo *repository.TaskRepository) {
	query := r.URL.Query()
	date := query.Get("date")

	if date != "" {
		// Обработка запроса с датой
		handleGetTasksByDate(w, date, repo)
	} else {
		// Обработка запроса без даты (с лимитом или без)
		handleGetTasksWithLimit(w, query.Get("limit"), repo)
	}
}

func handleGetTasksByDate(w http.ResponseWriter, date string, repo *repository.TaskRepository) {
	tasks, err := repo.GetByDate(date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.SendJSON(w, http.StatusOK, tasks)
}

func handleGetTasksWithLimit(w http.ResponseWriter, limitStr string, repo *repository.TaskRepository) {
	limit := DefaultTaskLimit

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			http.Error(w, "некорректное значение limit", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	tasks, err := repo.GetTasksWithLimit(limit)
	if err != nil {
		http.Error(w, "ошибка получения задач", http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []*models.Task{}
	}

	response := TasksResp{
		Tasks: tasks,
	}

	api.SendJSON(w, http.StatusOK, response)
}
