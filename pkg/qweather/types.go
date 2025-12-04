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

// DailyForecastResponse represents the response from QWeather API for daily forecast
type DailyForecastResponse struct {
	Code  string          `json:"code"`
	Daily []DailyForecast `json:"daily"`
}

// DailyForecast represents daily weather forecast data
type DailyForecast struct {
	FxDate         string `json:"fxDate"`         // Forecast date
	Sunrise        string `json:"sunrise"`        // Sunrise time
	Sunset         string `json:"sunset"`         // Sunset time
	Moonrise       string `json:"moonrise"`       // Moonrise time
	Moonset        string `json:"moonset"`        // Moonset time
	MoonPhase      string `json:"moonPhase"`      // Moon phase
	MoonPhaseIcon  string `json:"moonPhaseIcon"`  // Moon phase icon
	TempMax        string `json:"tempMax"`        // Maximum temperature
	TempMin        string `json:"tempMin"`        // Minimum temperature
	IconDay        string `json:"iconDay"`        // Daytime weather icon
	TextDay        string `json:"textDay"`        // Daytime weather description
	IconNight      string `json:"iconNight"`      // Nighttime weather icon
	TextNight      string `json:"textNight"`      // Nighttime weather description
	Wind360Day     string `json:"wind360Day"`     // Daytime wind direction in degrees
	WindDirDay     string `json:"windDirDay"`     // Daytime wind direction
	WindScaleDay   string `json:"windScaleDay"`   // Daytime wind scale
	WindSpeedDay   string `json:"windSpeedDay"`   // Daytime wind speed km/h
	Wind360Night   string `json:"wind360Night"`   // Nighttime wind direction in degrees
	WindDirNight   string `json:"windDirNight"`   // Nighttime wind direction
	WindScaleNight string `json:"windScaleNight"` // Nighttime wind scale
	WindSpeedNight string `json:"windSpeedNight"` // Nighttime wind speed km/h
	Humidity       string `json:"humidity"`       // Relative humidity
	Precip         string `json:"precip"`         // Precipitation amount mm
	Pressure       string `json:"pressure"`       // Atmospheric pressure hPa
	Vis            string `json:"vis"`            // Visibility km
	Cloud          string `json:"cloud"`          // Cloud cover percentage
	UvIndex        string `json:"uvIndex"`        // UV index
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

// AirNowResponse represents the response from QWeather API for current air quality
type AirNowResponse struct {
	Code string `json:"code"`
	Now  AirNow `json:"now"`
}

// AirNow represents current air quality data
type AirNow struct {
	PubTime  string `json:"pubTime"`  // Publication time
	Aqi      string `json:"aqi"`      // Air Quality Index
	Level    string `json:"level"`    // Air quality level
	Category string `json:"category"` // Air quality category
	Primary  string `json:"primary"`  // Primary pollutant
	Pm10     string `json:"pm10"`     // PM10 concentration
	Pm2p5    string `json:"pm2p5"`    // PM2.5 concentration
	No2      string `json:"no2"`      // NO2 concentration
	So2      string `json:"so2"`      // SO2 concentration
	Co       string `json:"co"`       // CO concentration
	O3       string `json:"o3"`       // O3 concentration
}

// AirQualityResponse represents the response from QWeather Air Quality API v1
type AirQualityResponse struct {
	Metadata   Metadata          `json:"metadata"`
	Indexes    []AirQualityIndex `json:"indexes"`
	Pollutants []Pollutant       `json:"pollutants"`
	Stations   []Station         `json:"stations"`
}

// Metadata represents response metadata
type Metadata struct {
	Tag string `json:"tag"`
}

// AirQualityIndex represents an air quality index (e.g., US EPA, QAQI)
type AirQualityIndex struct {
	Code             string           `json:"code"`             // Index code (e.g., "us-epa", "qaqi")
	Name             string           `json:"name"`             // Index name
	Aqi              float64          `json:"aqi"`              // AQI value
	AqiDisplay       string           `json:"aqiDisplay"`       // AQI display value
	Level            string           `json:"level"`            // Level
	Category         string           `json:"category"`         // Category (e.g., "Good")
	Color            Color            `json:"color"`            // Color code
	PrimaryPollutant PrimaryPollutant `json:"primaryPollutant"` // Primary pollutant
	Health           Health           `json:"health"`           // Health advice
}

// Color represents RGBA color
type Color struct {
	Red   int     `json:"red"`
	Green int     `json:"green"`
	Blue  int     `json:"blue"`
	Alpha float64 `json:"alpha"`
}

// PrimaryPollutant represents the primary pollutant
type PrimaryPollutant struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	FullName string `json:"fullName"`
}

// Health represents health effects and advice
type Health struct {
	Effect string       `json:"effect"`
	Advice HealthAdvice `json:"advice"`
}

// HealthAdvice represents health advice for different populations
type HealthAdvice struct {
	GeneralPopulation   string `json:"generalPopulation"`
	SensitivePopulation string `json:"sensitivePopulation"`
}

// Pollutant represents a specific pollutant's data
type Pollutant struct {
	Code          string        `json:"code"`          // Pollutant code (pm2p5, pm10, etc.)
	Name          string        `json:"name"`          // Pollutant name
	FullName      string        `json:"fullName"`      // Full name
	Concentration Concentration `json:"concentration"` // Concentration value
	SubIndexes    []SubIndex    `json:"subIndexes"`    // Sub-indexes
}

// Concentration represents pollutant concentration
type Concentration struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// SubIndex represents a sub-index for a pollutant
type SubIndex struct {
	Code       string  `json:"code"`
	Aqi        float64 `json:"aqi"`
	AqiDisplay string  `json:"aqiDisplay"`
}

// Station represents a monitoring station
type Station struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AirDailyResponse represents the response from QWeather API for daily air quality forecast
type AirDailyResponse struct {
	Code  string     `json:"code"`
	Daily []AirDaily `json:"daily"`
}

// AirDaily represents daily air quality forecast
type AirDaily struct {
	FxDate   string `json:"fxDate"`   // Forecast date
	Aqi      string `json:"aqi"`      // Air Quality Index
	Level    string `json:"level"`    // Air quality level
	Category string `json:"category"` // Air quality category
	Primary  string `json:"primary"`  // Primary pollutant
}

// WarningResponse represents the response from QWeather API for weather warnings
type WarningResponse struct {
	Code    string    `json:"code"`
	Warning []Warning `json:"warning"`
}

// Warning represents weather warning data
type Warning struct {
	ID            string `json:"id"`            // Warning ID
	Sender        string `json:"sender"`        // Issuing authority
	PubTime       string `json:"pubTime"`       // Publication time
	Title         string `json:"title"`         // Warning title
	StartTime     string `json:"startTime"`     // Warning start time
	EndTime       string `json:"endTime"`       // Warning end time
	Status        string `json:"status"`        // Warning status (active/update/cancel)
	Level         string `json:"level"`         // Warning level
	Severity      string `json:"severity"`      // Severity level
	SeverityColor string `json:"severityColor"` // Warning color code
	Type          string `json:"type"`          // Warning type code
	TypeName      string `json:"typeName"`      // Warning type name
	Text          string `json:"text"`          // Warning details
}
