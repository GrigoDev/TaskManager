package nextdate

import (
	"strconv"
	"strings"
	"time"
)

func NextDate(dateStr, repeat string) string {
	if repeat == "" {
		return ""
	}

	// Парсим дату
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return ""
	}

	// Разбиваем строку повтора на части
	parts := strings.Fields(repeat)
	if len(parts) < 2 {
		return ""
	}

	// Обрабатываем различные типы повторов
	switch parts[0] {
	case "y":
		// Ежегодно
		next := date.AddDate(1, 0, 0)
		// Обрабатываем 29 февраля
		if date.Month() == 2 && date.Day() == 29 {
			if !isLeap(next.Year()) {
				next = time.Date(next.Year(), 3, 1, 0, 0, 0, 0, time.UTC)
			}
		}
		return next.Format("20060102")

	case "d":
		// Ежедневно
		if len(parts) != 2 {
			return ""
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return ""
		}
		next := date.AddDate(0, 0, days)
		return next.Format("20060102")

	case "m":
		// Ежемесячно
		if len(parts) < 2 {
			return ""
		}
		// Парсим дни
		daysStr := strings.Split(parts[1], ",")
		var days []int
		for _, d := range daysStr {
			day, err := strconv.Atoi(d)
			if err != nil {
				return ""
			}
			days = append(days, day)
		}

		// Находим следующую дату
		next := date.AddDate(0, 1, 0)
		for {
			lastDay := daysInMonth(next.Year(), next.Month())
			for _, day := range days {
				if day > 0 {
					if day <= lastDay {
						return time.Date(next.Year(), next.Month(), day, 0, 0, 0, 0, time.UTC).Format("20060102")
					}
				} else {
					// Обрабатываем отрицательные дни (от конца месяца)
					actualDay := lastDay + day + 1
					if actualDay > 0 {
						return time.Date(next.Year(), next.Month(), actualDay, 0, 0, 0, 0, time.UTC).Format("20060102")
					}
				}
			}
			next = next.AddDate(0, 1, 0)
		}

	case "w":
		// Еженедельно
		if len(parts) < 2 {
			return ""
		}
		// Парсим дни недели
		weekdaysStr := strings.Split(parts[1], ",")
		var weekdays []int
		for _, w := range weekdaysStr {
			weekday, err := strconv.Atoi(w)
			if err != nil || weekday < 1 || weekday > 7 {
				return ""
			}
			weekdays = append(weekdays, weekday)
		}

		// Находим следующую дату
		next := date.AddDate(0, 0, 1)
		for i := 0; i < 7; i++ {
			weekday := int(next.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			for _, w := range weekdays {
				if weekday == w {
					return next.Format("20060102")
				}
			}
			next = next.AddDate(0, 0, 1)
		}
	}

	return ""
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
