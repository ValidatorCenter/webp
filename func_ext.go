package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
)

// Вывод служебного сообщения
func log(tp string, msg1 string, msg2 interface{}) {
	timeClr := fmt.Sprintf(color.MagentaString("[%s]"), time.Now().Format("2006-01-02 15:04:05"))
	msg0 := ""
	if tp == "ERR" {
		msg0 = fmt.Sprintf(color.RedString("ERROR: %s"), msg1)
	} else if tp == "WRN" {
		msg0 = fmt.Sprintf(color.HiYellowString("WARRNING: %s"), msg1)
	} else if tp == "INF" {
		infTag := fmt.Sprintf(color.YellowString("%s"), msg1)
		msg0 = fmt.Sprintf("%s: %#v", infTag, msg2)
	} else if tp == "OK" {
		msg0 = fmt.Sprintf(color.GreenString("%s"), msg1)
	} else if tp == "STR" {
		msg0 = fmt.Sprintf(color.CyanString("%s"), msg1)
	} else {
		msg0 = msg1
	}
	fmt.Printf("%s %s\n", timeClr, msg0)
}

// Сокращение строки
func getMinString(bigStr string) string {
	if len(bigStr) > 8 {
		return fmt.Sprintf("%s...%s", bigStr[:6], bigStr[len(bigStr)-4:len(bigStr)])
	} else {
		log("WRN", fmt.Sprint("getMinString(", bigStr, ")"), "")
		return bigStr
	}
}

// Вычисляет разницу между 2-мя датами, ответ в: годах, месяцах, днях, часах, минутах и секундах
func diffTime(a, b time.Time) (year, month, day, hour, min, sec int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}

// Вычисляет разницу между 2-мя датами, ответ в: строка
func diffTimeStr(dt1, dt2 time.Time) string {
	_Age := ""
	dYr, dMes, dDay, dHr, dMin, dSec := diffTime(dt1, dt2)
	if dYr > 0 {
		_Age = fmt.Sprintf("%dy %dm %dd", dYr, dMes, dDay)
	} else if dMes > 0 {
		_Age = fmt.Sprintf("%dm %dd %dh", dMes, dDay, dHr)
	} else if dDay > 0 {
		_Age = fmt.Sprintf("%dd %dh", dDay, dHr)
	} else if dHr > 0 {
		_Age = fmt.Sprintf("%dh %dm", dHr, dMin)
	} else if dMin > 0 {
		_Age = fmt.Sprintf("%dm %ds", dMin, dSec)
	} else if dSec > 0 {
		_Age = fmt.Sprintf("%ds", dSec)
	} else {
		_Age = "-"
	}
	return _Age
}

// Преобразует дату в строку нужного формата
func timeFormat0(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute())
}

// Конвертирование строки в число с точкой
func cnvStr2Float(amntTokenStr string) float32 {
	var fAmntToken float32 = 0.0
	if amntTokenStr != "" {
		fAmntToken64, err := strconv.ParseFloat(amntTokenStr, 64)
		if err != nil {
			panic(err.Error())
		}
		fAmntToken = float32(fAmntToken64)
	}
	return fAmntToken
}
