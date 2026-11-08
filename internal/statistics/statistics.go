package statistics

import (
	"database/sql"
	"fmt"
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
	if len(journals) == 0 {
		return 0
	}

	streak := 0
	checkDate := now

	// 오늘 일기 작성했는지 확인
	todayStr := now.Format("2006-01-02")
	hasToday := false
	for _, j := range journals {
		if j.Date == todayStr {
			hasToday = true
			break
		}
	}

	if !hasToday {
		// 오늘 작성하지 않았다면 어제부터 확인
		checkDate = now.AddDate(0, 0, -1)
	}

	for {
		dateStr := checkDate.Format("2006-01-02")
		found := false
		for _, j := range journals {
			if j.Date == dateStr {
				found = true
				break
			}
		}

		if found {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak
}

// 가장 긴 연속 작성 일수 계산
func calculateLongestStreak(journals []journal.Journal) int {
	if len(journals) == 0 {
		return 0
	}

	// 날짜를 집합으로 변환
	dateSet := make(map[string]bool)
	for _, j := range journals {
		dateSet[j.Date] = true
	}

	longest := 0
	current := 0

	// 가장 오래된 날짜부터 확인
	var earliest time.Time
	for _, j := range journals {
		parsed, err := time.Parse("2006-01-02", j.Date)
		if err == nil && (earliest.IsZero() || parsed.Before(earliest)) {
			earliest = parsed
		}
	}

	if !earliest.IsZero() {
		for d := earliest; ; d = d.AddDate(0, 0, 1) {
			if dateSet[d.Format("2006-01-02")] {
				current++
				if current > longest {
					longest = current
				}
			} else {
				current = 0
			}
			// 2년 뒤까지 확인
			if d.After(earliest.AddDate(2, 0, 0)) {
				break
			}
		}
	}

	return longest
}

// 통계 표시용 문자열 생성
func FormatStars(rating float64) string {
	fullStars := int(rating)
	emptyStars := 5 - fullStars
	if emptyStars < 0 {
		emptyStars = 0
	}
	return fmt.Sprintf("%s%s", 
		repeat('★', fullStars), 
		repeat('☆', emptyStars))
}

func repeat(r rune, count int) string {
	s := make([]rune, count)
	for i := range s {
		s[i] = r
	}
	return string(s)
}
