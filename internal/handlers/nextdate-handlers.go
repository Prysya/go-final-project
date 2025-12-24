package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Prysya/go-final-project/pkg/api"
	"github.com/Prysya/go-final-project/pkg/constants"
)

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeatStr := r.FormValue("repeat")

	var now time.Time
	if nowStr == "" {
		now = time.Now()
		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		var err error
		now, err = time.Parse(constants.DateFormat, nowStr)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte(""))
			return
		}
	}

	if dateStr == "" || repeatStr == "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(""))
		return
	}

	nextDate, err := api.NextDate(now, dateStr, repeatStr)
	if err != nil {
		fmt.Printf("Ошибка NextDate: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(""))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(nextDate))
}
