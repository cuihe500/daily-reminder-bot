package qweather

// WeatherResponse represents the response from QWeather API for current weather
type WeatherResponse struct {
	Code string         `json:"code"`
	Now  CurrentWeather `json:"now"`
}

// CurrentWeather represents current weather data
type CurrentWeather struct {
	Temp      string `json:"temp"`      // Temperature in Celsius
	FeelsLike string `json:"feelsLike"` // Feels like temperature
	Text      string `json:"text"`      // Weather description
	Humidity  string `json:"humidity"`  // Humidity percentage
	Wind360   string `json:"wind360"`   // Wind direction in degrees
	WindDir   string `json:"windDir"`   // Wind direction description
	WindScale string `json:"windScale"` // Wind scale
	WindSpeed string `json:"windSpeed"` // Wind speed km/h
}

// LifeIndicesResponse represents the response from QWeather API for life indices
type LifeIndicesResponse struct {
	Code  string      `json:"code"`
	Daily []LifeIndex `json:"daily"`
}

// LifeIndex represents a life index (e.g., clothing, UV, sports)
type LifeIndex struct {
	Type     string `json:"type"`     // Index type (1=sport, 3=dressing, 5=UV, etc.)
	Name     string `json:"name"`     // Index name
	Level    string `json:"level"`    // Level (e.g., "1", "2", "3")
	Category string `json:"category"` // Category description
	Text     string `json:"text"`     // Detailed advice
}

// GeoLocationResponse represents the response from QWeather GeoAPI
type GeoLocationResponse struct {
	Code     string        `json:"code"`
	Location []GeoLocation `json:"location"`
}

// GeoLocation represents a geographical location
type GeoLocation struct {
	Name      string `json:"name"`      // Location name
	ID        string `json:"id"`        // Location ID
	Lat       string `json:"lat"`       // Latitude
	Lon       string `json:"lon"`       // Longitude
	Adm2      string `json:"adm2"`      // Administrative division level 2 (district)
	Adm1      string `json:"adm1"`      // Administrative division level 1 (province/state)
	Country   string `json:"country"`   // Country
	Timezone  string `json:"tz"`        // Timezone
	UtcOffset string `json:"utcOffset"` // UTC offset
	Type      string `json:"type"`      // Location type
}
