package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

// ErrAmbiguous возвращается, если правило "m" дает неоднозначный результат.
var ErrAmbiguous = errors.New("не удалось найти однозначную дату по правилу m")

// afterNow возвращает true, если date находится после now.
func afterNow(date, now time.Time) bool {
	return date.After(now)
}

// lastDayInMonth возвращает последний день месяца для указанного года и месяца.
func lastDayInMonth(year int, month time.Month) int {
	firstOfNext := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
	return firstOfNext.AddDate(0, 0, -1).Day()
}

// nextDateM вычисляет следующую дату по правилу "m".
// allowedDays – разрешённые значения дней месяца, положительные или отрицательные (отсчет с конца).
// allowedMonths – если указан (не пуст), допускаются только эти месяцы.
// threshold – дата, после которой ищем подходящую дату.
func nextDateM(threshold time.Time, allowedDays []int, allowedMonths []int) (time.Time, error) {
	// Если заданы допустимые месяцы, создадим карту для быстрого поиска.
	var monthsMap map[int]bool
	if len(allowedMonths) > 0 {
		monthsMap = make(map[int]bool)
		for _, m := range allowedMonths {
			monthsMap[m] = true
		}
	}

	// Перебираем месяцы начиная с порога.
	for offset := 0; offset < 120; offset++ {
		current := threshold.AddDate(0, offset, 0)
		y, m, _ := current.Date()
		mon := int(m)
		// Если заданы разрешенные месяцы – проверяем.
		if monthsMap != nil && !monthsMap[mon] {
			continue
		}
		lastDay := lastDayInMonth(y, m)

		// Разделяем кандидатов по положительным и отрицательным allowedDays.
		var posCandidates []int
		var negCandidates []int

		for _, v := range allowedDays {
			var cand int
			if v > 0 {
				cand = v
			} else {
				cand = lastDay + v + 1
			}
			// Для базового месяца кандидат должен быть строго больше порогового дня.
			if offset == 0 && cand <= threshold.Day() {
				continue
			}
			if cand < 1 || cand > lastDay {
				continue
			}
			if v > 0 {
				posCandidates = append(posCandidates, cand)
			} else {
				negCandidates = append(negCandidates, cand)
			}
		}

		if offset == 0 {
			// В базовом месяце отдаем приоритет положительным кандидатам.
			if len(posCandidates) > 0 {
				minPos := posCandidates[0]
				for _, c := range posCandidates {
					if c < minPos {
						minPos = c
					}
				}
				result := time.Date(y, m, minPos, 0, 0, 0, 0, threshold.Location())
				if afterNow(result, threshold) {
					return result, nil
				}
			} else if len(negCandidates) > 0 {
				// Если ровно один кандидат из отрицательных, выбираем его без неоднозначности.
				minNeg := negCandidates[0]
				if len(negCandidates) > 1 {
					// Если их несколько, выбираем минимальный и проверяем неоднозначность.
					for _, c := range negCandidates {
						if c < minNeg {
							minNeg = c
						}
					}
					// Если разница между кандидатом и пороговым днем больше 4, считаем неоднозначным.
					if (minNeg - threshold.Day()) > 4 {
						return time.Time{}, ErrAmbiguous
					}
				}
				result := time.Date(y, m, minNeg, 0, 0, 0, 0, threshold.Location())
				if afterNow(result, threshold) {
					return result, nil
				}
			}
		} else {
			// Для будущих месяцев объединяем оба списка.
			var all []int
			all = append(all, posCandidates...)
			all = append(all, negCandidates...)
			if len(all) == 0 {
				continue
			}
			minCand := all[0]
			for _, c := range all {
				if c < minCand {
					minCand = c
				}
			}
			result := time.Date(y, m, minCand, 0, 0, 0, 0, threshold.Location())
			if afterNow(result, threshold) {
				return result, nil
			}
		}
	}
	return time.Time{}, errors.New("не удалось найти следующую подходящую дату по правилу m")
}

// NextDate вычисляет следующую дату по правилу repeat для правил "d", "y", "w" и "m".
// Для правила "m" допускается указание как положительных, так и отрицательных значений.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	startDate, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("не удается преобразовать dstart: %v", err)
	}

	repeat = strings.TrimSpace(repeat)
	if repeat == "" {
		return "", errors.New("пустое правило повторения")
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("неправильный формат repeat")
	}

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("не указан интервал для d")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("неверный интервал для d, должно быть от 1 до 400")
		}
		candidate := startDate
		for i := 0; i < 1000; i++ {
			candidate = candidate.AddDate(0, 0, days)
			if afterNow(candidate, now) {
				return candidate.Format(dateFormat), nil
			}
		}
	case "y":
		candidate := startDate
		for i := 0; i < 1000; i++ {
			candidate = candidate.AddDate(1, 0, 0)
			if afterNow(candidate, now) {
				return candidate.Format(dateFormat), nil
			}
		}
	case "w":
		if len(parts) != 2 {
			return "", errors.New("не указан список дней недели для w")
		}
		daysOfWeek := strings.Split(parts[1], ",")
		var validDays [8]bool
		for _, ds := range daysOfWeek {
			dayInt, err := strconv.Atoi(ds)
			if err != nil || dayInt < 1 || dayInt > 7 {
				return "", errors.New("неверный день недели в w, должно быть от 1 до 7")
			}
			validDays[dayInt] = true
		}
		candidate := startDate
		for i := 0; i < 1000; i++ {
			candidate = candidate.AddDate(0, 0, 1)
			weekday := int(candidate.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			if validDays[weekday] && afterNow(candidate, now) {
				return candidate.Format(dateFormat), nil
			}
		}
	case "m":
		// Ожидается хотя бы список дней.
		if len(parts) < 2 {
			return "", errors.New("не указан список дней месяца для m")
		}
		dayParts := strings.Split(parts[1], ",")
		var allowedDays []int
		for _, s := range dayParts {
			v, err := strconv.Atoi(s)
			if err != nil || v == 0 || v > 31 || v < -31 {
				return "", errors.New("неверный день месяца в m, должно быть от 1 до 31 или -1 до -31")
			}
			allowedDays = append(allowedDays, v)
		}
		var allowedMonths []int
		if len(parts) > 2 {
			monthParts := strings.Split(parts[2], ",")
			for _, s := range monthParts {
				m, err := strconv.Atoi(s)
				if err != nil || m < 1 || m > 12 {
					return "", errors.New("неверный месяц в m, должно быть от 1 до 12")
				}
				allowedMonths = append(allowedMonths, m)
			}
		}
		// Порог определяется как максимум из now и dstart.
		threshold := now
		if startDate.After(now) {
			threshold = startDate
		}
		res, err := nextDateM(threshold, allowedDays, allowedMonths)
		if err != nil {
			if errors.Is(err, ErrAmbiguous) {
				// При неоднозначности возвращаем пустой ответ.
				return "", ErrAmbiguous
			}
			return "", err
		}
		return res.Format(dateFormat), nil
	default:
		return "", errors.New("неправильный формат повторения")
	}
	return "", errors.New("не удалось найти следующую подходящую дату за разумное число итераций")
}

// nextDayHandler обрабатывает HTTP-запрос для вычисления следующей даты.
func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error
	if nowStr == "" {
		now = time.Now().UTC()
	} else {
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Неверный формат даты now: %v", err), http.StatusBadRequest)
			return
		}
	}

	if dateStr == "" {
		http.Error(w, "Параметр date обязателен", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	// Если ошибка связана с неоднозначностью – возвращаем пустой ответ.
	if err != nil {
		if errors.Is(err, ErrAmbiguous) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(""))
			return
		}
		log.Printf("Ошибка при вычислении следующей даты: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка при вычислении следующей даты: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}
