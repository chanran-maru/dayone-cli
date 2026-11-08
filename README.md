# DayOne CLI

터미널에서 사용하는 개인 생산성 도구입니다. 일정 관리, 일기 작성, 작업 추적, 일일 리뷰를 터미널에서 효율적으로 관리할 수 있습니다.

## 기능

- **📅 캘린더** (`calendar`): 현재 월의 캘린더를 표시합니다.
- **📝 일기** (`journal`): 일기를 작성, 조회, 수정, 삭제합니다.
- **✅ 작업 관리** (`task`): 작업을 추가, 완료, 삭제하고 진행 상황을 추적합니다.
- **📊 일일 리뷰** (`review`): 하루를 감정과 평점으로 기록하고 되돌아봅니다.
- **📈 통계** (`stats`): 일기, 작업, 리뷰의 통계를 확인합니다.
- **📋 오늘 요약** (`today`): 오늘의 일기, 작업, 리뷰를 한눈에 봅니다.

## 설치

```bash
go install .
```

## 사용법

### 캘린더

```bash
# 현재 월 캘린더 보기
DayOneCLI calendar
```

### 일기

```bash
# 오늘의 일기 작성
DayOneCLI journal write

# 특정 날짜 일기 읽기
DayOneCLI journal read 2026-11-08

# 모든 일기 목록 보기
DayOneCLI journal list

# 일기 수정
DayOneCLI journal edit 2026-11-08

# 일기 삭제
DayOneCLI journal delete 2026-11-08
```

### 작업 관리

```bash
# 작업 추가
DayOneCLI task add

# 모든 작업 목록
DayOneCLI task list

# 진행 중 작업
DayOneCLI task pending

# 완료된 작업
DayOneCLI task done

# 작업 완료
DayOneCLI task complete 1

# 작업 삭제
DayOneCLI task delete 1
```

### 일일 리뷰

```bash
# 오늘의 리뷰 작성
DayOneCLI review write

# 특정 날짜 리뷰 읽기
DayOneCLI review read 2026-11-08

# 모든 리뷰 목록
DayOneCLI review list

# 리뷰 삭제
DayOneCLI review delete 2026-11-08
```

### 통계

```bash
# 전체 통계
DayOneCLI stats

# 일기 통계
DayOneCLI stats journal

# 작업 통계
DayOneCLI stats tasks

# 리뷰 통계
DayOneCLI stats reviews
```

### 오늘 요약

```bash
# 오늘의 일기, 작업, 리뷰 확인
DayOneCLI today
```

## 기술 스택

- **Go** — 언어
- **Cobra** — CLI 프레임워크
- **Bubbletea** — TUI 프레임워크
- **SQLite** — 로컬 데이터베이스

## 라이선스

MIT
