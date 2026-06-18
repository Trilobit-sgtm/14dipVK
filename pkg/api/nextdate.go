package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату выполнения задачи.
//
//   - now    — точка отсчёта, результат должен быть строго больше неё
//   - dstart — исходная дата задачи в формате 20060102
//   - repeat — правило повторения (d <дни>, y, w <дни недели>, m <дни> [месяцы])
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}

	date, err := time.Parse(DateLayout, dstart)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %v", err)
	}

	parts := strings.SplitN(repeat, " ", 2)
	rule := parts[0]

	switch rule {
	case "y":
		if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
			return "", fmt.Errorf("правило 'y' не принимает дополнительные параметры")
		}
		return nextYear(date, now)

	case "d":
		if len(parts) < 2 {
			return "", fmt.Errorf("правило 'd' требует число дней: d <число>")
		}
		return nextDays(date, now, parts[1])

	case "w":
		if len(parts) < 2 {
			return "", fmt.Errorf("правило 'w' требует дни недели: w <1-7,...>")
		}
		return nextWeekdays(date, now, parts[1])

	case "m":
		if len(parts) < 2 {
			return "", fmt.Errorf("правило 'm' требует дни месяца: m <дни> [месяцы]")
		}
		return nextMonthdays(date, now, parts[1])

	default:
		return "", fmt.Errorf("неподдерживаемый формат правила: %s", repeat)
	}
}

// afterNow возвращает true, если date строго после now (сравниваем только дату, без времени).
func afterNow(date, now time.Time) bool {
	return dateOnly(date).After(dateOnly(now))
}

// nextYear — правило "y": двигаем дату по годам до ближайшей будущей.
func nextYear(date, now time.Time) (string, error) {
	for {
		date = date.AddDate(1, 0, 0)
		if afterNow(date, now) {
			return date.Format(DateLayout), nil
		}
	}
}

// nextDays — правило "d <число>": двигаем дату интервалами по N дней.
func nextDays(date, now time.Time, arg string) (string, error) {
	interval, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil || interval <= 0 || interval > 400 {
		return "", fmt.Errorf("неверный интервал для 'd': должно быть от 1 до 400")
	}

	for {
		date = date.AddDate(0, 0, interval)
		if afterNow(date, now) {
			return date.Format(DateLayout), nil
		}
	}
}

// nextWeekdays — правило "w <1-7,...>": ближайший подходящий день недели.
// 1 — понедельник, 7 — воскресенье.
func nextWeekdays(date, now time.Time, arg string) (string, error) {
	var weekday [8]bool // индексы 1..7

	for _, s := range strings.Split(arg, ",") {
		d, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || d < 1 || d > 7 {
			return "", fmt.Errorf("недопустимый день недели: %s (ожидается 1-7)", s)
		}
		weekday[d] = true
	}

	// быстрая перемотка: если dstart далеко в прошлом — прыгаем сразу
	// к now целыми неделями, сохраняя день недели
	if date.Before(now) {
		days := int(now.Sub(date).Hours() / 24)
		weeks := days / 7
		date = date.AddDate(0, 0, weeks*7)
	}

	// теперь ищем в пределах 14 дней — гарантированно найдём любой день недели
	for i := 0; i < 14; i++ {
		date = date.AddDate(0, 0, 1)
		wd := int(date.Weekday())
		if wd == 0 {
			wd = 7 // воскресенье: Go=0, нам нужно 7
		}
		if weekday[wd] && afterNow(date, now) {
			return date.Format(DateLayout), nil
		}
	}

	return "", fmt.Errorf("не удалось найти подходящий день недели")
}

// nextMonthdays — правило "m <дни> [месяцы]": ближайший подходящий день месяца.
// Допускаются -1 (последний день) и -2 (предпоследний).
func nextMonthdays(date, now time.Time, arg string) (string, error) {
	// разбиваем на дни и (опционально) месяцы
	subParts := strings.SplitN(arg, " ", 2)

	var day [32]bool    // индексы 1..31 + спецзначения обрабатываем отдельно
	var specDay [3]bool // 0 — не используем, 1 — -1, 2 — -2
	var month [13]bool  // индексы 1..12; если ни один не указан, подходит любой

	anyMonth := true

	// парсим дни
	for _, s := range strings.Split(subParts[0], ",") {
		d, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return "", fmt.Errorf("некорректный день месяца: %s", s)
		}
		switch {
		case d == -1:
			specDay[1] = true
		case d == -2:
			specDay[2] = true
		case d >= 1 && d <= 31:
			day[d] = true
		default:
			return "", fmt.Errorf("недопустимый день месяца: %d (ожидается 1-31, -1, -2)", d)
		}
	}

	// парсим месяцы (если указаны)
	if len(subParts) == 2 {
		anyMonth = false
		for _, s := range strings.Split(subParts[1], ",") {
			m, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || m < 1 || m > 12 {
				return "", fmt.Errorf("недопустимый месяц: %s (ожидается 1-12)", s)
			}
			month[m] = true
		}
	}

	// перебираем дни начиная с date+1
	date = date.AddDate(0, 0, 1)
	for i := 0; i < 366*2; i++ {
		m := int(date.Month())
		if anyMonth || month[m] {
			d := date.Day()
			// последний день месяца
			lastDay := lastDayOf(date)
			if (specDay[1] && d == lastDay) ||
				(specDay[2] && d == lastDay-1) ||
				(d <= 31 && day[d]) {
				if afterNow(date, now) {
					return date.Format(DateLayout), nil
				}
			}
		}
		date = date.AddDate(0, 0, 1)
	}

	return "", fmt.Errorf("не удалось найти подходящий день месяца")
}

// lastDayOf возвращает номер последнего дня в месяце переданной даты.
func lastDayOf(t time.Time) int {
	// первый день следующего месяца минус 1 день
	first := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	return first.AddDate(0, 0, -1).Day()
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

// nextDateHandler обрабатывает GET /api/nextdate?now=...&date=...&repeat=...
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(DateLayout, nowStr)
		if err != nil {
			http.Error(w, "некорректный параметр now", http.StatusBadRequest)
			return
		}
	}

	result, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(result))
}
