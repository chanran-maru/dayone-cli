package task

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Task struct {
	ID        int
	Title     string
	Status    string // pending, completed
	CreatedAt string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CREATE: 작업 추가
func (r *Repository) Create(title string) error {
	if title == "" {
		return fmt.Errorf("작업 제목은 비워둘 수 없습니다")
	}

	now := time.Now().Format(time.RFC3339)

	query := `
		INSERT INTO tasks
			(title, status, created_at)
		VALUES (?, 'pending', ?)
	`

	_, err := r.db.Exec(query, title, now)
	return err
}

// READ: 작업 ID로 조회
func (r *Repository) GetByID(id int) (Task, error) {
	var t Task

	query := `
		SELECT id, title, status, created_at
		FROM tasks
		WHERE id = ?
	`

	err := r.db.QueryRow(query, id).Scan(
		&t.ID,
		&t.Title,
		&t.Status,
		&t.CreatedAt,
	)

	return t, err
}

// READ: 상태별 작업 목록
func (r *Repository) List(status string) ([]Task, error) {
	query := `
		SELECT id, title, status, created_at
		FROM tasks
		WHERE status = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task

		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Status,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// READ: 전체 작업 목록
func (r *Repository) ListAll() ([]Task, error) {
	query := `
		SELECT id, title, status, created_at
		FROM tasks
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task

		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Status,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// UPDATE: 작업 제목 수정
func (r *Repository) Update(id int, newTitle string) error {
	if newTitle == "" {
		return fmt.Errorf("작업 제목은 비워둘 수 없습니다")
	}

	result, err := r.db.Exec(`
		UPDATE tasks
		SET title = ?
		WHERE id = ?
	`, newTitle, id)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UPDATE: 작업 완료
func (r *Repository) Complete(id int) error {
	result, err := r.db.Exec(`
		UPDATE tasks
		SET status = 'completed'
		WHERE id = ? AND status = 'pending'
	`, id)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DELETE: 작업 삭제
func (r *Repository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// READ: 날짜별 작업 목록
func (r *Repository) ListByDate(date string) ([]Task, error) {
	start, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil || start.Format("2006-01-02") != date {
		return nil, fmt.Errorf("날짜는 YYYY-MM-DD 형식이어야 합니다")
	}
	end := start.AddDate(0, 0, 1)

	query := `
		SELECT id, title, status, created_at
		FROM tasks
		WHERE julianday(created_at) >= julianday(?)
		  AND julianday(created_at) < julianday(?)
		ORDER BY created_at DESC
	`

	// RFC3339의 시간대 오프셋을 반영해 로컬 날짜의 자정부터 다음 자정 전까지 조회한다.
	rows, err := r.db.Query(query, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task

		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Status,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
