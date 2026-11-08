package review

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Review struct {
	Date      string
	Mood      string
	Rating    int
	Content   string
	CreatedAt string
	UpdatedAt string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// 날짜 형식 검증
func ValidateDate(date string) error {
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("날짜는 YYYY-MM-DD 형식이어야 합니다")
	}
	return nil
}

// 감정 검증
func ValidateMood(mood string) error {
	validMoods := map[string]bool{
		"😊": true, "😐": true, "😢": true, "😡": true, "😴": true,
		"happy": true, "neutral": true, "sad": true, "angry": true, "tired": true,
	}
	if !validMoods[mood] {
		return fmt.Errorf("감정은 다음 중 하나여야 합니다: 😊, 😐, 😢, 😡, 😴 (또는 happy, neutral, sad, angry, tired)")
	}
	return nil
}

// 평점 검증
func ValidateRating(rating int) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("평점은 1~5 사이여야 합니다")
	}
	return nil
}

// CREATE: 일일 리뷰 작성 (이미 존재하면 업데이트)
func (r *Repository) Create(date, mood string, rating int, content string) error {
	if err := ValidateDate(date); err != nil {
		return err
	}
	if err := ValidateMood(mood); err != nil {
		return err
	}
	if err := ValidateRating(rating); err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)

	query := `
		INSERT INTO reviews
			(date, mood, rating, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE
			SET mood = excluded.mood, rating = excluded.rating,
				content = excluded.content, updated_at = excluded.updated_at
	`

	_, err := r.db.Exec(query, date, mood, rating, content, now, now)
	return err
}

// READ: 특정 날짜 리뷰 조회
func (r *Repository) GetByDate(date string) (Review, error) {
	if err := ValidateDate(date); err != nil {
		return Review{}, err
	}

	var rev Review

	query := `
		SELECT date, mood, rating, content, created_at, updated_at
		FROM reviews
		WHERE date = ?
	`

	err := r.db.QueryRow(query, date).Scan(
		&rev.Date,
		&rev.Mood,
		&rev.Rating,
		&rev.Content,
		&rev.CreatedAt,
		&rev.UpdatedAt,
	)

	return rev, err
}

// READ: 전체 리뷰 목록
func (r *Repository) List() ([]Review, error) {
	query := `
		SELECT date, mood, rating, content, created_at, updated_at
		FROM reviews
		ORDER BY date DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review

	for rows.Next() {
		var rev Review

		if err := rows.Scan(
			&rev.Date,
			&rev.Mood,
			&rev.Rating,
			&rev.Content,
			&rev.CreatedAt,
			&rev.UpdatedAt,
		); err != nil {
			return nil, err
		}

		reviews = append(reviews, rev)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

// DELETE: 리뷰 삭제
func (r *Repository) Delete(date string) error {
	if err := ValidateDate(date); err != nil {
		return err
	}

	result, err := r.db.Exec("DELETE FROM reviews WHERE date = ?", date)
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

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
