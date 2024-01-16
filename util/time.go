package util

import (
	"fmt"
	"strings"
	"time"
)

/*
 * time tools
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//macro define
const (
	//seconds
	OneMinSec = 60
	OneHourSec = OneMinSec * 60
	OneDaySec = OneHourSec * 24
	OneMonthSec = OneDaySec * 30
	OneYearSec = OneMonthSec * 12

	//general
	TimeLayoutStr = "2006-01-02 15:04:05"
)

//face info
type Time struct {
	controlDuration time.Duration
}

//get current utc time
func (f *Time) Now() time.Time {
	if f.controlDuration != 0 {
		return time.Now().Add(f.controlDuration).UTC()
	}
	return time.Now().UTC()
}

//reset server control duration
func (f *Time) ResetControlDuration() {
	f.controlDuration = 0
}

//set server control duration
func (f *Time) SetControlDuration(
	duration time.Duration) {
	f.controlDuration = duration
}

//change server control duration
func (f *Time) ChangeControlDuration(
	timeStr string) (time.Duration, error) {
	nowT := time.Now()
	changeTime, err := time.Parse(TimeLayoutStr, timeStr)
	if err != nil {
		return 0, err
	}
	f.controlDuration = changeTime.Sub(nowT)
	return f.controlDuration, nil
}

//begin of time period
func (f *Time) BeginningOfDay() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
func (f *Time) BeginningOfWeek() time.Time {
	t := f.BeginningOfDay()
	weekday := int(t.Weekday())
	weekStartDay := int(time.Monday)

	if weekday < weekStartDay {
		weekday = weekday + 7 - weekStartDay
	} else {
		weekday = weekday - weekStartDay
	}
	return t.AddDate(0, 0, -weekday)
}
func (f *Time) BeginningOfMonth() time.Time {
	y, m, _ := time.Now().Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
}

//convert string format time to unix
func (f *Time) ConvertStrTime2Unix(timeStr string) (int64, error) {
	nowT := time.Now()
	tm, err := time.Parse(TimeLayoutStr, timeStr)
	if err != nil {
		return 0, err
	}
	tm.Sub(nowT)
	return tm.UTC().Unix(), nil
}

//convert timestamp to date format string
func (f *Time) Timestamp2Str(timeStamp int64) string {
	return time.Unix(timeStamp, 0).UTC().Format(TimeLayoutStr)
}

//convert date time string to timestamp
func (f *Time) DateTime2Unix(dateTime string) (int64, error) {
	//remove un useful info
	dateTime = strings.Replace(dateTime, "T", " ", -1)
	dateTime = strings.Replace(dateTime, "Z", "", -1)

	//theTime, err := time.Parse(TimeLayOut, dateTime)
	theTime, err := time.ParseInLocation(TimeLayoutStr, dateTime, time.Local)
	if err != nil {
		return 0, err
	}
	return theTime.Unix(), nil
}

//get current date, like YYYY-MM-DD
func (f *Time) GetCurDate() string {
	now := time.Now()
	curDate := fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day())
	return curDate
}