package repository

import (
	"database/sql"
	"fmt"

	"github.com/Prysya/go-final-project/pkg/db"
	"github.com/Prysya/go-final-project/pkg/models"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository() (*TaskRepository, error) {
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("база данных не инициализирована")
	}
	return &TaskRepository{db: database}, nil
}

func (r *TaskRepository) Create(task *models.Task) (int64, error) {
	query := `insert into scheduler (date, title, comment, repeat) values (?, ?, ?, ?)`

	result, err := r.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания задачи: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID созданной задачи: %w", err)
	}

	return id, nil
}

func (r *TaskRepository) GetByID(id int) (*models.Task, error) {
	query := `select id, date, title, comment, repeat from scheduler where id = ?`

	row := r.db.QueryRow(query, id)

	var task models.Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения задачи: %w", err)
	}

	return &task, nil
}

func (r *TaskRepository) GetAll() ([]models.Task, error) {
	query := `select id, date, title, comment, repeat from scheduler order by date`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения задач: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования задачи: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по задачам: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) GetByDate(date string) ([]models.Task, error) {
	if !models.IsValidDate(date) {
		return nil, fmt.Errorf("некорректный формат даты: %s", date)
	}

	query := `select id, date, title, comment, repeat from scheduler where date = ? order by id`

	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения задач по дате: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования задачи: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по задачам: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) Update(task *models.Task) error {
	query := `update scheduler set date = ?, title = ?, comment = ?, repeat = ? where id = ?`

	result, err := r.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("ошибка обновления задачи: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки обновления: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача с ID %d не найдена", task.ID)
	}

	return nil
}

func (r *TaskRepository) Delete(id int) error {
	query := `delete from scheduler where id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления задачи: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки удаления: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача с ID %d не найдена", id)
	}

	return nil
}

func (r *TaskRepository) DeleteByDate(date string) (int64, error) {
	if !models.IsValidDate(date) {
		return 0, fmt.Errorf("некорректный формат даты: %s", date)
	}

	query := `delete from scheduler where date = ?`

	result, err := r.db.Exec(query, date)
	if err != nil {
		return 0, fmt.Errorf("ошибка удаления задач по дате: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("ошибка проверки удаления: %w", err)
	}

	return rowsAffected, nil
}

func (r *TaskRepository) Migrate() error {
	query := `select name from sqlite_master where type='table' and name='scheduler'`

	var tableName string
	err := r.db.QueryRow(query).Scan(&tableName)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("ошибка проверки таблицы: %w", err)
	}

	return nil
}
