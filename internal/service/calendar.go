package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/pkg/calendar"
	"github.com/cuichanghe/daily-reminder-bot/pkg/holiday"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
)

// CalendarService provides calendar-related functionality
type CalendarService struct {
	calculator    *calendar.Calculator
	holidayClient *holiday.Client
	timezone      *time.Location
}

// NewCalendarService creates a new CalendarService
func NewCalendarService(timezone *time.Location, holidayClient *holiday.Client) *CalendarService {
	return &CalendarService{
		calculator:    calendar.NewCalculator(timezone),
		holidayClient: holidayClient,
		timezone:      timezone,
	}
}

// FormatDateHeader formats the date header with both solar and lunar dates
// Example: 今天是 2025年1月28日 农历甲辰年腊月廿九
func (s *CalendarService) FormatDateHeader(date time.Time) string {
	logger.Debug("FormatDateHeader called",
		zap.Time("date", date))

	info := s.calculator.GetDateInfo(date)

	// Handle leap month
	monthStr := info.LunarMonthCN
	if info.IsLeapMonth {
		monthStr = "闰" + monthStr
	}

	result := fmt.Sprintf("今天是 %d年%d月%d日 农历%s%s%s",
		date.Year(), int(date.Month()), date.Day(),
		info.LunarYearCN, monthStr, info.LunarDayCN)

	logger.Debug("Date header formatted",
		zap.String("lunar_year", info.LunarYearCN),
		zap.String("lunar_month", monthStr),
		zap.String("lunar_day", info.LunarDayCN))

	return result
}

// FormatTodaySpecial formats today's special dates (festivals/solar terms)
// Returns empty string if no special dates
func (s *CalendarService) FormatTodaySpecial(date time.Time) string {
	logger.Debug("FormatTodaySpecial called", zap.Time("date", date))

	var specials []string

	// Check today's solar term
	jieQi := s.calculator.GetTodayJieQi(date)
	if jieQi != "" {
		specials = append(specials, jieQi)
		logger.Debug("Today's solar term found", zap.String("jie_qi", jieQi))
	}

	// Check today's festivals
	festivals := s.calculator.GetTodayFestivals(date)
	specials = append(specials, festivals...)
	if len(festivals) > 0 {
		logger.Debug("Today's festivals found", zap.Strings("festivals", festivals))
	}

	if len(specials) == 0 {
		logger.Debug("No special dates today")
		return ""
	}

	logger.Debug("Special dates formatted",
		zap.Int("count", len(specials)),
		zap.Strings("specials", specials))
	return fmt.Sprintf("【%s】", strings.Join(specials, " | "))
}

// FormatUpcomingFestivals formats the upcoming festivals countdown
func (s *CalendarService) FormatUpcomingFestivals(date time.Time, limit int) string {
	logger.Debug("FormatUpcomingFestivals called",
		zap.Time("date", date),
		zap.Int("limit", limit))

	festivals := s.calculator.GetUpcomingFestivals(date, limit+5) // Get extra for filtering

	if len(festivals) == 0 {
		logger.Debug("No upcoming festivals found")
		return ""
	}

	logger.Debug("Upcoming festivals retrieved",
		zap.Int("count", len(festivals)))

	// Try to get statutory holiday info from API for accurate holiday days
	var nextStatutory *holiday.StatutoryHoliday
	if s.holidayClient != nil {
		var err error
		nextStatutory, err = s.holidayClient.GetNextHoliday(date)
		if err != nil {
			logger.Warn("Failed to get next statutory holiday",
				zap.Error(err))
		} else if nextStatutory != nil {
			logger.Debug("Next statutory holiday retrieved",
				zap.String("name", nextStatutory.Name),
				zap.Int("days_until", nextStatutory.DaysUntil))
		}
	}

	var builder strings.Builder
	builder.WriteString("📅 近期节日/节气：\n")

	count := 0
	for _, f := range festivals {
		if count >= limit {
			break
		}

		emoji := f.Type.Emoji()

		// Check if this is the statutory holiday from API and update holiday days
		holidayDays := f.HolidayDays
		if nextStatutory != nil && f.Name == nextStatutory.Name && f.IsHoliday {
			// Use API data if available (more accurate)
			return ""
		}

		if f.DaysUntil == 0 {
			// Today
			if f.IsHoliday && holidayDays > 0 {
				builder.WriteString(fmt.Sprintf("%s 今天是%s！（放假%d天）\n",
					emoji, f.Name, holidayDays))
			} else {
				builder.WriteString(fmt.Sprintf("%s 今天是%s！\n", emoji, f.Name))
			}
		} else {
			// Future
			if f.IsHoliday && holidayDays > 0 {
				builder.WriteString(fmt.Sprintf("%s 还有%d天到%s（放假%d天）\n",
					emoji, f.DaysUntil, f.Name, holidayDays))
			} else {
				builder.WriteString(fmt.Sprintf("%s 还有%d天到%s\n",
					emoji, f.DaysUntil, f.Name))
			}
		}

		count++
	}

	return builder.String()
}

// GetCalendarInfo returns comprehensive calendar information for AI prompts
func (s *CalendarService) GetCalendarInfo(date time.Time) *calendar.CalendarInfo {
	logger.Debug("GetCalendarInfo called", zap.Time("date", date))

	info := s.calculator.GetDateInfo(date)
	festivals := s.calculator.GetUpcomingFestivals(date, 5)
	todayFestivals := s.calculator.GetTodayFestivals(date)
	todayJieQi := s.calculator.GetTodayJieQi(date)

	logger.Debug("Calendar info retrieved",
		zap.Int("upcoming_festivals", len(festivals)),
		zap.Int("today_festivals", len(todayFestivals)),
		zap.String("today_jie_qi", todayJieQi))

	return &calendar.CalendarInfo{
		DateInfo:          info,
		UpcomingFestivals: festivals,
		TodayFestivals:    todayFestivals,
		TodayJieQi:        todayJieQi,
	}
}

// FormatCalendarInfoForAI formats calendar information for AI prompts
func (s *CalendarService) FormatCalendarInfoForAI(date time.Time) string {
	logger.Debug("FormatCalendarInfoForAI called", zap.Time("date", date))

	info := s.GetCalendarInfo(date)
	if info == nil || info.DateInfo == nil {
		logger.Debug("No calendar info available")
		return ""
	}

	var builder strings.Builder

	// Date info
	builder.WriteString(fmt.Sprintf("公历: %d年%d月%d日\n",
		date.Year(), int(date.Month()), date.Day()))
	builder.WriteString(fmt.Sprintf("农历: %s%s%s\n",
		info.DateInfo.LunarYearCN, info.DateInfo.LunarMonthCN, info.DateInfo.LunarDayCN))
	builder.WriteString(fmt.Sprintf("生肖: %s\n", info.DateInfo.Zodiac))

	// Today's special
	if info.TodayJieQi != "" {
		builder.WriteString(fmt.Sprintf("今日节气: %s\n", info.TodayJieQi))
	}
	if len(info.TodayFestivals) > 0 {
		builder.WriteString(fmt.Sprintf("今日节日: %s\n", strings.Join(info.TodayFestivals, ", ")))
	}

	// Upcoming festivals
	if len(info.UpcomingFestivals) > 0 {
		builder.WriteString("近期节日:\n")
		for _, f := range info.UpcomingFestivals {
			if f.DaysUntil > 0 {
				builder.WriteString(fmt.Sprintf("- %s（%d天后）\n", f.Name, f.DaysUntil))
			}
		}
	}

	return builder.String()
}
