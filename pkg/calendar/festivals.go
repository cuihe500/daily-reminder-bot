package calendar

import "time"

// SolarFestival represents a solar calendar festival with fixed date
type SolarFestival struct {
	Month int
	Day   int
	Name  string
	Type  FestivalType
}

// LunarFestival represents a lunar calendar festival
type LunarFestival struct {
	Month int
	Day   int
	Name  string
	Type  FestivalType
}

// FloatingFestival represents a festival with a floating date
type FloatingFestival struct {
	Name       string
	Type       FestivalType
	Calculator func(year int) time.Time
}

// SolarFestivals contains all fixed-date solar festivals
var SolarFestivals = []SolarFestival{
	// 中国公历节日
	{1, 1, "元旦", FestivalTypeStatutory},
	{3, 8, "妇女节", FestivalTypeSolar},
	{3, 12, "植树节", FestivalTypeSolar},
	{5, 1, "劳动节", FestivalTypeStatutory},
	{5, 4, "青年节", FestivalTypeSolar},
	{6, 1, "儿童节", FestivalTypeSolar},
	{7, 1, "建党节", FestivalTypeSolar},
	{8, 1, "建军节", FestivalTypeSolar},
	{9, 10, "教师节", FestivalTypeSolar},
	{10, 1, "国庆节", FestivalTypeStatutory},

	// 西方节日
	{2, 14, "情人节", FestivalTypeWestern},
	{4, 1, "愚人节", FestivalTypeWestern},
	{10, 31, "万圣节", FestivalTypeWestern},
	{12, 25, "圣诞节", FestivalTypeWestern},
}

// LunarFestivals contains all lunar calendar festivals
var LunarFestivals = []LunarFestival{
	{1, 1, "春节", FestivalTypeStatutory},
	{1, 15, "元宵节", FestivalTypeLunar},
	{2, 2, "龙抬头", FestivalTypeLunar},
	{5, 5, "端午节", FestivalTypeStatutory},
	{7, 7, "七夕节", FestivalTypeLunar},
	{7, 15, "中元节", FestivalTypeLunar},
	{8, 15, "中秋节", FestivalTypeStatutory},
	{9, 9, "重阳节", FestivalTypeLunar},
	{12, 8, "腊八节", FestivalTypeLunar},
	{12, 23, "小年", FestivalTypeLunar},
	// 除夕需要特殊处理（腊月最后一天）
}

// FloatingFestivals contains festivals with variable dates
var FloatingFestivals = []FloatingFestival{
	{
		Name: "母亲节",
		Type: FestivalTypeFloating,
		Calculator: func(year int) time.Time {
			// 5月第2个周日
			return getNthWeekday(year, time.May, time.Sunday, 2)
		},
	},
	{
		Name: "父亲节",
		Type: FestivalTypeFloating,
		Calculator: func(year int) time.Time {
			// 6月第3个周日
			return getNthWeekday(year, time.June, time.Sunday, 3)
		},
	},
	{
		Name: "感恩节",
		Type: FestivalTypeFloating,
		Calculator: func(year int) time.Time {
			// 11月第4个周四（美国）
			return getNthWeekday(year, time.November, time.Thursday, 4)
		},
	},
	{
		Name: "复活节",
		Type: FestivalTypeFloating,
		Calculator: func(year int) time.Time {
			return calculateEaster(year)
		},
	},
}

// getNthWeekday calculates the nth occurrence of a weekday in a given month
func getNthWeekday(year int, month time.Month, weekday time.Weekday, n int) time.Time {
	// Start from the first day of the month
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	// Find the first occurrence of the weekday
	daysUntilWeekday := int(weekday - firstDay.Weekday())
	if daysUntilWeekday < 0 {
		daysUntilWeekday += 7
	}
	firstOccurrence := firstDay.AddDate(0, 0, daysUntilWeekday)

	// Calculate the nth occurrence
	return firstOccurrence.AddDate(0, 0, (n-1)*7)
}

// calculateEaster calculates Easter Sunday using the Anonymous Gregorian algorithm
func calculateEaster(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
