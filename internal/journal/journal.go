package journal

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Journal struct {
	Date      string
	Content   string
	CreatedAt string
	UpdatedAt string
}

type Repository struct {
	db *sql.DB
}

// Repository 생성
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// 날짜 형식 검증
func ValidateDate(date string) error {
	parsed, err := time.Parse("2006-01-02", date)

	if err != nil || parsed.Format("2006-01-02") != date {
		return fmt.Errorf("날짜는 YYYY-MM-DD 형식이어야 합니다")
	}

	return nil
}

// CREATE: 일기 작성 (이미 존재하면 업데이트)
func (r *Repository) Create(date, content string) error {
	if err := ValidateDate(date); err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)

	query := `
        INSERT INTO journals
            (date, content, created_at, updated_at)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(date) DO UPDATE
            SET content = excluded.content, updated_at = excluded.updated_at
    `

	_, err := r.db.Exec(query, date, content, now, now)
	return err
}

// READ: 특정 날짜 일기 조회
func (r *Repository) GetByDate(date string) (Journal, error) {
	if err := ValidateDate(date); err != nil {
		return Journal{}, err
	}

	var j Journal

	query := `
        SELECT date, content, created_at, updated_at
        FROM journals
        WHERE date = ?
    `

	err := r.db.QueryRow(query, date).Scan(
		&j.Date,
		&j.Content,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	return j, err
}

// READ: 전체 일기 목록
func (r *Repository) List() ([]Journal, error) {
	query := `
        SELECT date, content, created_at, updated_at
        FROM journals
        ORDER BY date DESC
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var journals []Journal

	for rows.Next() {
		var j Journal

		if err := rows.Scan(
			&j.Date,
			&j.Content,
			&j.CreatedAt,
			&j.UpdatedAt,
		); err != nil {
			return nil, err
		}

		journals = append(journals, j)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return journals, nil
}

// UPDATE: 일기 수정
func (r *Repository) Update(date, content string) error {
	if err := ValidateDate(date); err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)

	result, err := r.db.Exec(`
        UPDATE journals
        SET content = ?, updated_at = ?
        WHERE date = ?
    `, content, now, date)

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

// DELETE: 일기 삭제
func (r *Repository) Delete(date string) error {
	if err := ValidateDate(date); err != nil {
		return err
	}

	result, err := r.db.Exec(
		"DELETE FROM journals WHERE date = ?", date,
	)
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

// 이미 존재하는 일기인지 확인할 때 사용
func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
