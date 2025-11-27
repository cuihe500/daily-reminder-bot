package calendar

import "time"

// DateInfo contains date information including solar and lunar calendars
type DateInfo struct {
	Solar        time.Time
	LunarYear    int
	LunarMonth   int
	LunarDay     int
	LunarYearCN  string // 甲辰年
	LunarMonthCN string // 腊月
	LunarDayCN   string // 初二
	IsLeapMonth  bool
	Zodiac       string // 龙
	GanZhi       string // 甲辰
}

// FestivalType represents the type of festival
type FestivalType int

const (
	FestivalTypeSolarTerm  FestivalType = iota + 1 // 节气
	FestivalTypeLunar                              // 农历节日
	FestivalTypeSolar                              // 公历节日
	FestivalTypeStatutory                          // 法定节假日
	FestivalTypeWestern                            // 西方节日
	FestivalTypeFloating                           // 浮动节日（如母亲节）
)

// String returns the Chinese name of the festival type
func (t FestivalType) String() string {
	switch t {
	case FestivalTypeSolarTerm:
		return "节气"
	case FestivalTypeLunar:
		return "农历"
	case FestivalTypeSolar:
		return "公历"
	case FestivalTypeStatutory:
		return "法定"
	case FestivalTypeWestern:
		return "西方"
	case FestivalTypeFloating:
		return "浮动"
	default:
		return "未知"
	}
}

// Emoji returns the emoji for the festival type
func (t FestivalType) Emoji() string {
	switch t {
	case FestivalTypeSolarTerm:
		return "🌿"
	case FestivalTypeLunar:
		return "🏮"
	case FestivalTypeSolar:
		return "📆"
	case FestivalTypeStatutory:
		return "🎉"
	case FestivalTypeWestern:
		return "🌍"
	case FestivalTypeFloating:
		return "💐"
	default:
		return "📌"
	}
}

// Festival represents a festival or solar term
type Festival struct {
	Name        string
	Date        time.Time
	Type        FestivalType
	DaysUntil   int
	IsHoliday   bool
	HolidayDays int
}

// CalendarInfo contains comprehensive calendar information
type CalendarInfo struct {
	DateInfo          *DateInfo
	UpcomingFestivals []Festival
	TodayFestivals    []string
	TodayJieQi        string
}
