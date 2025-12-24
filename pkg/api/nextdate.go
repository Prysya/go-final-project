package api

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Prysya/go-final-project/pkg/constants"
)

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("пустое правило повторения")
	}

	startDate, err := time.Parse(constants.DateFormat, dstart)
	if err != nil {
		return "", errors.New("неверный формат начальной даты")
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("неверный формат правила")
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "d":
		return handleDailyRule(now, startDate, args)
	case "y":
		return handleYearlyRule(now, startDate, args)
	case "w":
		return handleWeeklyRule(now, startDate, args)
	case "m":
		return handleMonthlyRule(now, startDate, args)
	default:
		return "", errors.New("неподдерживаемый формат правила")
	}
}

func handleDailyRule(now, startDate time.Time, args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.New("неверный формат для правила 'd'")
	}

	interval, err := strconv.Atoi(args[0])
	if err != nil {
		return "", errors.New("неверный интервал дней")
	}

	if interval <= 0 || interval > 400 {
		return "", errors.New("интервал дней должен быть от 1 до 400")
	}

	date := startDate

	for {
		date = date.AddDate(0, 0, interval)
		if afterNow(date, now) {
			break
		}
	}

	return formatDate(date), nil
}

func handleYearlyRule(now, startDate time.Time, args []string) (string, error) {
	if len(args) != 0 {
		return "", errors.New("правило 'y' не принимает аргументов")
	}

	date := startDate

	for {
		date = date.AddDate(1, 0, 0)

		if date.Day() == 29 && date.Month() == 2 {
			nextDay := time.Date(date.Year(), 2, 29, 0, 0, 0, 0, date.Location())
			if nextDay.Day() != 29 {
				date = time.Date(date.Year(), 3, 1, 0, 0, 0, 0, date.Location())
			}
		}

		if afterNow(date, now) {
			break
		}
	}

	return formatDate(date), nil
}

func handleWeeklyRule(now, startDate time.Time, args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.New("неверный формат для правила 'w'")
	}

	daysStr := strings.Split(args[0], ",")
	weekdays := make(map[int]bool)

	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 1 || day > 7 {
			return "", errors.New("дни недели должны быть от 1 до 7")
		}
		weekdays[day] = true
	}

	if len(weekdays) == 0 {
		return "", errors.New("не указаны дни недели")
	}

	date := startDate

	if isValidWeekday(date, weekdays) && afterNow(date, now) {
		return formatDate(date), nil
	}

	searchDate := startDate
	if now.After(startDate) {
		searchDate = now
	}

	searchDate = searchDate.AddDate(0, 0, 1)

	for i := 0; i < 366; i++ {
		if isValidWeekday(searchDate, weekdays) {
			return formatDate(searchDate), nil
		}
		searchDate = searchDate.AddDate(0, 0, 1)
	}

	return "", errors.New("не удалось найти подходящую дату")
}

func isValidWeekday(date time.Time, weekdays map[int]bool) bool {
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return weekdays[weekday]
}

func handleMonthlyRule(now, startDate time.Time, args []string) (string, error) {
	if len(args) < 1 || len(args) > 2 {
		return "", errors.New("неверный формат для правила 'm'")
	}

	daysStr := strings.Split(args[0], ",")
	dayRules := make([]int, 0, len(daysStr))

	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return "", errors.New("неверный формат дня месяца")
		}
		if (day < -2 || day == 0 || day > 31) && day != -1 {
			return "", errors.New("день месяца должен быть от -2 до 31, кроме 0")
		}
		dayRules = append(dayRules, day)
	}

	monthRules := make(map[int]bool)
	if len(args) == 2 {
		monthsStr := strings.Split(args[1], ",")
		for _, monthStr := range monthsStr {
			month, err := strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > 12 {
				return "", errors.New("месяцы должны быть от 1 до 12")
			}
			monthRules[month] = true
		}
	}

	if isValidMonthlyDate(startDate, dayRules, monthRules) && afterNow(startDate, now) {
		return formatDate(startDate), nil
	}

	searchDate := startDate
	if now.After(startDate) {
		searchDate = now
	}

	searchDate = searchDate.AddDate(0, 0, 1)

	for i := 0; i < 365*5; i++ {
		if isValidMonthlyDate(searchDate, dayRules, monthRules) && afterNow(searchDate, now) {
			return formatDate(searchDate), nil
		}
		searchDate = searchDate.AddDate(0, 0, 1)
	}

	return "", errors.New("не удалось найти подходящую дату")
}

func isValidMonthlyDate(date time.Time, dayRules []int, monthRules map[int]bool) bool {
	if len(monthRules) > 0 {
		month := int(date.Month())
		if !monthRules[month] {
			return false
		}
	}

	year, month, _ := date.Date()
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	preLastDay := lastDay - 1

	for _, dayRule := range dayRules {
		switch dayRule {
		case -1:
			if date.Day() == lastDay {
				return true
			}
		case -2:
			if date.Day() == preLastDay {
				return true
			}
		default:
			if date.Day() == dayRule {
				if dayRule <= lastDay {
					return true
				}
			}
		}
	}

	return false
}

func afterNow(date, now time.Time) bool {
	dateYear, dateMonth, dateDay := date.Date()
	nowYear, nowMonth, nowDay := now.Date()

	return dateYear > nowYear ||
		(dateYear == nowYear && dateMonth > nowMonth) ||
		(dateYear == nowYear && dateMonth == nowMonth && dateDay > nowDay)
}

func formatDate(date time.Time) string {
	return date.Format(constants.DateFormat)
}
