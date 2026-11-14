package statistics

import (
	"database/sql"
	"strings"
	"time"

	"DayOneCLI/internal/journal"
	"DayOneCLI/internal/review"
	"DayOneCLI/internal/task"
)

type StatisticsService struct {
	journalRepo *journal.Repository
	taskRepo    *task.Repository
	reviewRepo  *review.Repository
}

func NewService(db *sql.DB) *StatisticsService {
	return &StatisticsService{
		journalRepo: journal.NewRepository(db),
		taskRepo:    task.NewRepository(db),
		reviewRepo:  review.NewRepository(db),
	}
}

// 일기 통계
type JournalStats struct {
	TotalCount    int
	ThisMonth     int
	ThisWeek      int
	CurrentStreak int
	LongestStreak int
}

// 작업 통계
type TaskStats struct {
	Total     int
	Pending   int
	Completed int
}

// 리뷰 통계
type ReviewStats struct {
	TotalCount int
	AvgRating  float64
	MoodCounts map[string]int
}

// 일기 통계 계산
func (s *StatisticsService) GetJournalStats() (JournalStats, error) {
	var stats JournalStats

	// 전체 일기 수
	journals, err := s.journalRepo.List()
	if err != nil {
		return stats, err
	}
	stats.TotalCount = len(journals)

	now := time.Now()
	thisMonth := now.Format("2006-01")
	thisWeekStart := now.AddDate(0, 0, -int(now.Weekday())).Format("2006-01-02")

	// 이번 달과 이번 주 계산
	for _, j := range journals {
		if len(j.Date) >= 7 && j.Date[:7] == thisMonth {
			stats.ThisMonth++
		}
		if j.Date >= thisWeekStart {
			stats.ThisWeek++
		}
	}

	// 연속 작성 일수 계산
	stats.CurrentStreak = calculateStreak(journals, now)
	stats.LongestStreak = calculateLongestStreak(journals)

	return stats, nil
}

// 작업 통계 계산
func (s *StatisticsService) GetTaskStats() (TaskStats, error) {
	var stats TaskStats

	pending, err := s.taskRepo.List("pending")
	if err != nil {
		return stats, err
	}
	stats.Pending = len(pending)

	completed, err := s.taskRepo.List("completed")
	if err != nil {
		return stats, err
	}
	stats.Completed = len(completed)

	stats.Total = stats.Pending + stats.Completed

	return stats, nil
}

// 리뷰 통계 계산
func (s *StatisticsService) GetReviewStats() (ReviewStats, error) {
	var stats ReviewStats
	stats.MoodCounts = make(map[string]int)

	reviews, err := s.reviewRepo.List()
	if err != nil {
		return stats, err
	}

	stats.TotalCount = len(reviews)

	var totalRating int
	for _, rev := range reviews {
		totalRating += rev.Rating
		stats.MoodCounts[rev.Mood]++
	}

	if stats.TotalCount > 0 {
		stats.AvgRating = float64(totalRating) / float64(stats.TotalCount)
	}

	return stats, nil
}

// 연속 작성 일수 계산
func calculateStreak(journals []journal.Journal, now time.Time) int {
	dateSet := make(map[string]bool, len(journals))
	for _, j := range journals {
		dateSet[j.Date] = true
	}

	streak := 0
	checkDate := now

	// 오늘 일기 작성했는지 확인
	if !dateSet[now.Format("2006-01-02")] {
		// 오늘 작성하지 않았다면 어제부터 확인
		checkDate = now.AddDate(0, 0, -1)
	}

	for dateSet[checkDate.Format("2006-01-02")] {
		streak++
		checkDate = checkDate.AddDate(0, 0, -1)
	}

	return streak
}

// 가장 긴 연속 작성 일수 계산
func calculateLongestStreak(journals []journal.Journal) int {
	dateSet := make(map[time.Time]bool, len(journals))
	for _, j := range journals {
		if date, err := time.Parse("2006-01-02", j.Date); err == nil {
			dateSet[date] = true
		}
	}

	longest := 0
	for date := range dateSet {
		// 각 연속 구간의 시작에서만 탐색하여 전체 날짜를 한 번씩 확인한다.
		if dateSet[date.AddDate(0, 0, -1)] {
			continue
		}
		current := 0
		for d := date; dateSet[d]; d = d.AddDate(0, 0, 1) {
			current++
		}
		if current > longest {
			longest = current
		}
	}

	return longest
}

// 통계 표시용 문자열 생성
func FormatStars(rating float64) string {
	// 범위를 벗어나거나 NaN인 값도 별 다섯 개로 안전하게 표시한다.
	if !(rating >= 0) {
		rating = 0
	}
	if rating > 5 {
		rating = 5
	}
	fullStars := int(rating)
	return strings.Repeat("★", fullStars) + strings.Repeat("☆", 5-fullStars)
}
