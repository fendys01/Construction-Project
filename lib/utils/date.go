package utils

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// ToUTCfromGMT7 ...
func ToUTCfromGMT7(strTime string) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Now(), err
	}

	date, err := time.ParseInLocation("2006-01-02 15:04:05", strTime, location)
	if err != nil {
		fmt.Printf("\nerror when parse strTime [%s] -> err: %v\n", strTime, err)
		return time.Now(), err
	}

	return date.In(time.UTC), nil
}

// FromUTCLocationToGMT7 ...
func FromUTCLocationToGMT7(date time.Time) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Now(), err
	}

	return date.In(location), nil
}

// FromGMT7LocationUTCMin7 ...
func FromGMT7LocationUTCMin7(date time.Time) (time.Time, error) {
	date = date.Add(time.Hour * -7)
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Now(), err
	}

	date = date.In(location)

	return date.In(time.UTC), nil
}

// TimeElapsed ...
func TimeElapsed(param time.Time) string {
	var text string
	var parts []string

	now := time.Now()

	currentYear, currentMonth, currentDay := now.Date()
	currentHour, currentMinute, currentSecond := now.Clock()

	paramYear, paramMonth, paramDay := param.Date()
	paramHour, paramMinute, paramSecond := param.Clock()

	year := math.Abs(float64(int(currentYear - paramYear)))
	month := math.Abs(float64(int(currentMonth - paramMonth)))
	day := math.Abs(float64(int(currentDay - paramDay)))
	hour := math.Abs(float64(int(currentHour - paramHour)))
	minute := math.Abs(float64(int(currentMinute - paramMinute)))
	second := math.Abs(float64(int(currentSecond - paramSecond)))
	week := math.Floor(day / 7)

	s := func(x float64) string {
		if int(x) == 1 {
			return ""
		}
		return "s"
	}

	if year > 0 {
		parts = append(parts, strconv.Itoa(int(year))+" Year"+s(year))
	}
	if month > 0 {
		parts = append(parts, strconv.Itoa(int(month))+" Month"+s(month))
	}
	if week > 0 {
		parts = append(parts, strconv.Itoa(int(week))+" Week"+s(week))
	}
	if day > 0 {
		parts = append(parts, strconv.Itoa(int(day))+" Day"+s(day))
	}
	if hour > 0 {
		parts = append(parts, strconv.Itoa(int(hour))+" Hour"+s(hour))
	}
	if minute > 0 {
		parts = append(parts, strconv.Itoa(int(minute))+" Minute"+s(minute))
	}
	if second > 0 {
		parts = append(parts, strconv.Itoa(int(second))+" Second"+s(second))
	}
	if len(parts) == 0 {
		return "Now"
	}
	if now.After(param) {
		text = " Ago"
	} else {
		text = " After"
	}

	return parts[0] + text
}
