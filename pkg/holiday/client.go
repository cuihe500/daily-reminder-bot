package holiday

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// StatutoryHoliday represents a statutory holiday with vacation days
type StatutoryHoliday struct {
	Name        string    `json:"name"`
	Date        time.Time `json:"date"`
	DaysUntil   int       `json:"rest"`
	HolidayDays int       `json:"holiday_days"`
	IsHoliday   bool      `json:"holiday"`
}

// Client is a Holiday API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	cache      map[string]*cacheEntry
	cacheMu    sync.RWMutex
	cacheTTL   time.Duration
}

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

// APIResponse represents the API response structure
type APIResponse struct {
	Code    int              `json:"code"`
	Holiday *HolidayData     `json:"holiday"`
	Type    *HolidayTypeData `json:"type"`
}

// HolidayData represents holiday information from the API
type HolidayData struct {
	Holiday bool   `json:"holiday"`
	Name    string `json:"name"`
	Wage    int    `json:"wage"`
	Date    string `json:"date"`
	Rest    int    `json:"rest"`
	After   *int   `json:"after"`
	Target  string `json:"target"`
}

// HolidayTypeData represents holiday type information
type HolidayTypeData struct {
	Type int    `json:"type"` // 0=工作日, 1=周末, 2=节日, 3=调休放假, 4=补班
	Name string `json:"name"`
	Week int    `json:"week"`
}

// NextHolidayResponse represents the response for next holiday API
type NextHolidayResponse struct {
	Code    int          `json:"code"`
	Holiday *HolidayData `json:"holiday"`
	Workday *HolidayData `json:"workday"`
}

// YearHolidaysResponse represents the response for year holidays API
type YearHolidaysResponse struct {
	Code    int                     `json:"code"`
	Holiday map[string]*HolidayData `json:"holiday"`
}

// NewClient creates a new Holiday API client
func NewClient(baseURL string, cacheTTL time.Duration) *Client {
	if cacheTTL == 0 {
		cacheTTL = 24 * time.Hour
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cache:      make(map[string]*cacheEntry),
		cacheTTL:   cacheTTL,
	}
}

// GetNextHoliday retrieves the next statutory holiday from a given date
func (c *Client) GetNextHoliday(date time.Time) (*StatutoryHoliday, error) {
	dateStr := date.Format("2006-01-02")
	cacheKey := fmt.Sprintf("next_%s", dateStr)

	// Check cache
	if cached := c.getFromCache(cacheKey); cached != nil {
		if h, ok := cached.(*StatutoryHoliday); ok {
			return h, nil
		}
	}

	url := fmt.Sprintf("%s/api/holiday/next/%s", c.baseURL, dateStr)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get next holiday: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var apiResp NextHolidayResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Code != 0 || apiResp.Holiday == nil {
		return nil, fmt.Errorf("API returned error code: %d", apiResp.Code)
	}

	holidayDate, _ := time.Parse("2006-01-02", apiResp.Holiday.Date)
	holiday := &StatutoryHoliday{
		Name:      apiResp.Holiday.Name,
		Date:      holidayDate,
		DaysUntil: apiResp.Holiday.Rest,
		IsHoliday: apiResp.Holiday.Holiday,
	}

	// Cache the result
	c.setCache(cacheKey, holiday)

	return holiday, nil
}

// GetYearHolidays retrieves all statutory holidays for a given year
func (c *Client) GetYearHolidays(year int) ([]StatutoryHoliday, error) {
	cacheKey := fmt.Sprintf("year_%d", year)

	// Check cache
	if cached := c.getFromCache(cacheKey); cached != nil {
		if h, ok := cached.([]StatutoryHoliday); ok {
			return h, nil
		}
	}

	url := fmt.Sprintf("%s/api/holiday/year/%d", c.baseURL, year)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get year holidays: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var apiResp YearHolidaysResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API returned error code: %d", apiResp.Code)
	}

	var holidays []StatutoryHoliday
	for _, h := range apiResp.Holiday {
		if h == nil || !h.Holiday {
			continue
		}
		holidayDate, _ := time.Parse("2006-01-02", h.Date)
		holidays = append(holidays, StatutoryHoliday{
			Name:      h.Name,
			Date:      holidayDate,
			DaysUntil: h.Rest,
			IsHoliday: h.Holiday,
		})
	}

	// Cache the result
	c.setCache(cacheKey, holidays)

	return holidays, nil
}

// GetDateInfo retrieves holiday information for a specific date
func (c *Client) GetDateInfo(date time.Time) (*HolidayData, *HolidayTypeData, error) {
	dateStr := date.Format("2006-01-02")
	url := fmt.Sprintf("%s/api/holiday/info/%s", c.baseURL, dateStr)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get date info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, nil, fmt.Errorf("API returned error code: %d", apiResp.Code)
	}

	return apiResp.Holiday, apiResp.Type, nil
}

func (c *Client) getFromCache(key string) interface{} {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	entry, ok := c.cache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil
	}
	return entry.data
}

func (c *Client) setCache(key string, data interface{}) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	c.cache[key] = &cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(c.cacheTTL),
	}
}
