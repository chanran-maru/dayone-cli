# DayOne CLI

터미널에서 사용하는 개인 일기 관리 도구입니다.

## 기능

- **일기 작성** (`journal write`): 오늘의 일기를 작성합니다.
- **일기 읽기** (`journal read [date]`): 특정 날짜의 일기를 조회합니다.
- **일기 목록** (`journal list`): 모든 일기를 날짜순으로 보여줍니다.
- **일기 수정** (`journal edit [date]`): 기존 일기를 수정합니다.
- **일기 삭제** (`journal delete [date]`): 일기를 삭제합니다.
- **캘린더** (`calendar`): 현재 월의 캘린더를 표시합니다.
- **오늘 요약** (`today`): 오늘의 요약을 보여줍니다.

## 설치

```bash
go install .
```

## 사용법

```bash
# 오늘의 일기 작성
DayOneCLI journal write

# 특정 날짜 일기 읽기
DayOneCLI journal read 2026-11-08

# 모든 일기 목록 보기
DayOneCLI journal list

# 캘린더 보기
DayOneCLI calendar
```

## 기술 스택

- **Go** — 언어
- **Cobra** — CLI 프레임워크
- **Bubbletea** — TUI 프레임워크
- **SQLite** — 로컬 데이터베이스

## 라이선스

MIT
