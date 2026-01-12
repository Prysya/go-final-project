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

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	repo, err := repository.NewTaskRepository()
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if r.URL.Query().Get("id") != "" {
			handleGetTaskByID(w, r, repo)
		} else {
			handleGetTasks(w, r, repo)
		}
	case http.MethodPost:
		handleCreateTask(w, r, repo)
	case http.MethodPut:
		handleUpdateTask(w, r, repo)
	case http.MethodDelete:
		handleDeleteTask(w, r, repo)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	repo, err := repository.NewTaskRepository()
	if err != nil {
		api.WriteJSONError(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	handleTaskDone(w, r, repo)
}

func handleTaskDone(w http.ResponseWriter, r *http.Request, repo *repository.TaskRepository) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		api.WriteJSONError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.WriteJSONError(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	task, err := repo.GetByID(id)
	if err != nil {
		api.WriteJSONError(w, "Ошибка получения задачи", http.StatusInternalServerError)
		return
	}

	if task == nil {
		api.WriteJSONError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	if task.Repeat == "" {
		err = repo.Delete(id)
		if err != nil {
			api.WriteJSONError(w, "Ошибка удаления задачи: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		now := time.Now()
		nextDate, err := api.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			api.WriteJSONError(w, "Ошибка расчета следующей даты: "+err.Error(), http.StatusInternalServerError)
			return
		}

		task.Date = nextDate
		err = repo.Update(task)
		if err != nil {
			api.WriteJSONError(w, "Ошибка обновления даты задачи: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	api.SendJSON(w, http.StatusOK, map[string]interface{}{})
}

func handleGetTaskByID(w http.ResponseWriter, r *http.Request, repo *repository.TaskRepository) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		api.WriteJSONError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.WriteJSONError(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	task, err := repo.GetByID(id)
	if err != nil {
		api.WriteJSONError(w, "Ошибка получения задачи", http.StatusInternalServerError)
		return
	}

	if task == nil {
		api.WriteJSONError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	api.SendJSON(w, http.StatusOK, task)

}

func handleGetTasks(w http.ResponseWriter, r *http.Request, repo *repository.TaskRepository) {
	date := r.URL.Query().Get("date")

	var tasks []models.Task
	var err error

	if date != "" {
		tasks, err = repo.GetByDate(date)
	} else {
		api.WriteJSONError(w, "Не указан параметр запроса date", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.SendJSON(w, http.StatusOK, tasks)
}

func handleCreateTask(w http.ResponseWriter, r *http.Request, repo *repository.TaskRepository) {
	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		api.WriteJSONError(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		api.WriteJSONError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	now := time.Now()
	today := now.Format(constants.DateFormat)

	if task.Date == "" {
		task.Date = today
	} else {
		_, err := time.Parse(constants.DateFormat, task.Date)
		if err != nil {
			api.WriteJSONError(w, fmt.Sprintf("Дата представлена в формате, отличном от %d", constants.DateFormat), http.StatusBadRequest)
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
				api.WriteJSONError(w, "Правило повторения указано в неправильном формате", http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	} else {
		if task.Repeat != "" {
			_, err := api.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				api.WriteJSONError(w, "Правило повторения указано в неправильном формате", http.StatusBadRequest)
				return
			}
		}
	}

	id, err := repo.Create(&task)
	if err != nil {
		api.WriteJSONError(w, "Ошибка при создании задачи: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id": strconv.FormatInt(id, 10),
	}

	api.SendJSON(w, http.StatusOK, response)
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request, repo *repository.TaskRepository) {
	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		api.WriteJSONError(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		api.WriteJSONError(w, "Не указан идентификатор задачи", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		api.WriteJSONError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(task.ID)
	if err != nil {
		api.WriteJSONError(w, "Некорректный идентификатор задачи", http.StatusBadRequest)
		return
	}

	existingTask, err := repo.GetByID(id)
	if err != nil {
		api.WriteJSONError(w, "Ошибка проверки задачи", http.StatusInternalServerError)
		return
	}
	if existingTask == nil {
		api.WriteJSONError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	now := time.Now()
	today := now.Format(constants.DateFormat)

	if task.Date == "" {
		task.Date = today
	} else {
		_, err := time.Parse(constants.DateFormat, task.Date)
		if err != nil {
			api.WriteJSONError(w, fmt.Sprintf("Дата представлена в формате, отличном от %s", constants.DateFormat), http.StatusBadRequest)
			return
		}
	}

	taskTime, _ := time.Parse(constants.DateFormat, task.Date)
	if taskTime.Before(now.Truncate(24*time.Hour)) && task.Repeat == "" {
		task.Date = today
	}

	if task.Repeat != "" {
		_, err := api.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			api.WriteJSONError(w, "Правило повторения указано в неправильном формате", http.StatusBadRequest)
			return
		}
	}

	err = repo.Update(&task)
	if err != nil {
		if err.Error() == fmt.Sprintf("задача с ID %d не найдена", id) {
			api.WriteJSONError(w, "Задача не найдена", http.StatusNotFound)
		} else {
			api.WriteJSONError(w, "Ошибка при обновлении задачи: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{}

	api.SendJSON(w, http.StatusOK, response)
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request, repo *repository.TaskRepository) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		api.WriteJSONError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.WriteJSONError(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	err = repo.Delete(id)
	if err != nil {
		if err.Error() == fmt.Sprintf("задача с ID %d не найдена", id) {
			api.WriteJSONError(w, "Задача не найдена", http.StatusNotFound)
		} else {
			api.WriteJSONError(w, "Ошибка удаления задачи: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	api.SendJSON(w, http.StatusOK, map[string]interface{}{})
}
