package calendar

import (
	"sort"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// Calculator handles date calculations for calendar information
type Calculator struct {
	timezone *time.Location
}

// NewCalculator creates a new Calculator with the specified timezone
func NewCalculator(timezone *time.Location) *Calculator {
	if timezone == nil {
		timezone = time.UTC
	}
	return &Calculator{timezone: timezone}
}

// GetDateInfo returns detailed date information for a given date
func (c *Calculator) GetDateInfo(date time.Time) *DateInfo {
	date = date.In(c.timezone)
	solar := calendar.NewSolarFromYmd(date.Year(), int(date.Month()), date.Day())
	lunar := solar.GetLunar()

	return &DateInfo{
		Solar:        date,
		LunarYear:    lunar.GetYear(),
		LunarMonth:   lunar.GetMonth(),
		LunarDay:     lunar.GetDay(),
		LunarYearCN:  lunar.GetYearInGanZhi() + "年",
		LunarMonthCN: lunar.GetMonthInChinese() + "月",
		LunarDayCN:   lunar.GetDayInChinese(),
		IsLeapMonth:  lunar.GetMonth() < 0,
		Zodiac:       lunar.GetYearShengXiao(),
		GanZhi:       lunar.GetYearInGanZhi(),
	}
}

// GetTodayJieQi returns the solar term for today, or empty string if none
func (c *Calculator) GetTodayJieQi(date time.Time) string {
	date = date.In(c.timezone)
	solar := calendar.NewSolarFromYmd(date.Year(), int(date.Month()), date.Day())
	lunar := solar.GetLunar()
	return lunar.GetJieQi()
}

// GetTodayFestivals returns a list of festivals for the given date
func (c *Calculator) GetTodayFestivals(date time.Time) []string {
	date = date.In(c.timezone)
	solar := calendar.NewSolarFromYmd(date.Year(), int(date.Month()), date.Day())
	lunar := solar.GetLunar()

	var festivals []string

	// Get lunar festivals from the library
	lunarFestivals := lunar.GetFestivals()
	for i := lunarFestivals.Front(); i != nil; i = i.Next() {
		festivals = append(festivals, i.Value.(string))
	}

	// Get other traditional festivals
	otherFestivals := lunar.GetOtherFestivals()
	for i := otherFestivals.Front(); i != nil; i = i.Next() {
		festivals = append(festivals, i.Value.(string))
	}

	// Check fixed solar festivals
	for _, sf := range SolarFestivals {
		if sf.Month == int(date.Month()) && sf.Day == date.Day() {
			festivals = append(festivals, sf.Name)
		}
	}

	// Check floating festivals
	for _, ff := range FloatingFestivals {
		festivalDate := ff.Calculator(date.Year())
		if festivalDate.Year() == date.Year() &&
			festivalDate.Month() == date.Month() &&
			festivalDate.Day() == date.Day() {
			festivals = append(festivals, ff.Name)
		}
	}

	return festivals
}

// GetUpcomingFestivals returns the upcoming festivals sorted by date
func (c *Calculator) GetUpcomingFestivals(date time.Time, limit int) []Festival {
	date = date.In(c.timezone)
	today := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, c.timezone)

	var festivals []Festival

	// Add solar festivals for current and next year
	festivals = append(festivals, c.getSolarFestivals(date)...)

	// Add lunar festivals
	festivals = append(festivals, c.getLunarFestivals(date)...)

	// Add floating festivals
	festivals = append(festivals, c.getFloatingFestivals(date)...)

	// Add solar terms
	festivals = append(festivals, c.getSolarTerms(date)...)

	// Filter to only include today and future dates, calculate DaysUntil
	var upcoming []Festival
	for _, f := range festivals {
		fDate := time.Date(f.Date.Year(), f.Date.Month(), f.Date.Day(), 0, 0, 0, 0, c.timezone)
		if !fDate.Before(today) {
			f.DaysUntil = int(fDate.Sub(today).Hours() / 24)
			upcoming = append(upcoming, f)
		}
	}

	// Sort by date
	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].Date.Before(upcoming[j].Date)
	})

	// Remove duplicates (same date and name)
	upcoming = removeDuplicates(upcoming)

	// Limit results
	if len(upcoming) > limit {
		upcoming = upcoming[:limit]
	}

	return upcoming
}

func (c *Calculator) getSolarFestivals(date time.Time) []Festival {
	var festivals []Festival
	years := []int{date.Year(), date.Year() + 1}

	for _, year := range years {
		for _, sf := range SolarFestivals {
			fDate := time.Date(year, time.Month(sf.Month), sf.Day, 0, 0, 0, 0, c.timezone)
			festivals = append(festivals, Festival{
				Name:      sf.Name,
				Date:      fDate,
				Type:      sf.Type,
				IsHoliday: sf.Type == FestivalTypeStatutory,
			})
		}
	}

	return festivals
}

func (c *Calculator) getLunarFestivals(date time.Time) []Festival {
	var festivals []Festival

	solar := calendar.NewSolarFromYmd(date.Year(), int(date.Month()), date.Day())
	lunar := solar.GetLunar()
	lunarYears := []int{lunar.GetYear(), lunar.GetYear() + 1}

	for _, year := range lunarYears {
		for _, lf := range LunarFestivals {
			lunarDate := calendar.NewLunarFromYmd(year, lf.Month, lf.Day)
			solarDate := lunarDate.GetSolar()

			fDate := time.Date(
				solarDate.GetYear(),
				time.Month(solarDate.GetMonth()),
				solarDate.GetDay(),
				0, 0, 0, 0, c.timezone,
			)

			festivals = append(festivals, Festival{
				Name:      lf.Name,
				Date:      fDate,
				Type:      lf.Type,
				IsHoliday: lf.Type == FestivalTypeStatutory,
			})
		}

		// Handle 除夕 (New Year's Eve) - last day of lunar year
		chuxi := c.getChuxi(year)
		if !chuxi.IsZero() {
			festivals = append(festivals, Festival{
				Name: "除夕",
				Date: chuxi,
				Type: FestivalTypeLunar,
			})
		}
	}

	return festivals
}

func (c *Calculator) getChuxi(lunarYear int) time.Time {
	// 除夕 is the last day of the 12th lunar month
	// Try 腊月三十 first
	lunarDate := calendar.NewLunarFromYmd(lunarYear, 12, 30)
	if lunarDate.GetMonth() == 12 && lunarDate.GetDay() == 30 {
		solarDate := lunarDate.GetSolar()
		return time.Date(
			solarDate.GetYear(),
			time.Month(solarDate.GetMonth()),
			solarDate.GetDay(),
			0, 0, 0, 0, c.timezone,
		)
	}

	// If 腊月 doesn't have 30 days, use 腊月二十九
	lunarDate = calendar.NewLunarFromYmd(lunarYear, 12, 29)
	solarDate := lunarDate.GetSolar()
	return time.Date(
		solarDate.GetYear(),
		time.Month(solarDate.GetMonth()),
		solarDate.GetDay(),
		0, 0, 0, 0, c.timezone,
	)
}

func (c *Calculator) getFloatingFestivals(date time.Time) []Festival {
	var festivals []Festival
	years := []int{date.Year(), date.Year() + 1}

	for _, year := range years {
		for _, ff := range FloatingFestivals {
			fDate := ff.Calculator(year)
			fDate = time.Date(fDate.Year(), fDate.Month(), fDate.Day(), 0, 0, 0, 0, c.timezone)
			festivals = append(festivals, Festival{
				Name: ff.Name,
				Date: fDate,
				Type: ff.Type,
			})
		}
	}

	return festivals
}

func (c *Calculator) getSolarTerms(date time.Time) []Festival {
	var festivals []Festival

	solar := calendar.NewSolarFromYmd(date.Year(), int(date.Month()), date.Day())
	lunar := solar.GetLunar()

	// Get the JieQi table which contains solar terms and their dates
	jieQiTable := lunar.GetJieQiTable()
	jieQiList := lunar.GetJieQiList()

	for i := jieQiList.Front(); i != nil; i = i.Next() {
		name := i.Value.(string)
		jqSolar := jieQiTable[name]
		if jqSolar == nil {
			continue
		}

		fDate := time.Date(
			jqSolar.GetYear(),
			time.Month(jqSolar.GetMonth()),
			jqSolar.GetDay(),
			0, 0, 0, 0, c.timezone,
		)

		f := Festival{
			Name: name,
			Date: fDate,
			Type: FestivalTypeSolarTerm,
		}

		// 清明 is a statutory holiday
		if name == "清明" {
			f.Type = FestivalTypeStatutory
			f.IsHoliday = true
		}

		festivals = append(festivals, f)
	}

	return festivals
}

func removeDuplicates(festivals []Festival) []Festival {
	seen := make(map[string]bool)
	var result []Festival

	for _, f := range festivals {
		key := f.Date.Format("2006-01-02") + "_" + f.Name
		if !seen[key] {
			seen[key] = true
			result = append(result, f)
		}
	}

	return result
}
