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

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	repo, err := repository.NewTaskRepository()
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetTasksWithLimit(w, r, repo)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleGetTasksWithLimit(w http.ResponseWriter, r *http.Request, repo *repository.TaskRepository) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50

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
