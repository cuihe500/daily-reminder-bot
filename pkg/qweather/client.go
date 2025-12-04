package qweather

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
)

// Client is a QWeather API client
type Client struct {
	authMode   string             // "jwt" or "api_key"
	apiKey     string             // API Key (for api_key mode)
	privateKey ed25519.PrivateKey // Ed25519 private key (for jwt mode)
	keyID      string             // Key ID (for jwt mode)
	projectID  string             // Project ID (for jwt mode)
	baseURL    string
	client     *http.Client
}

// NewClient creates a new QWeather API client with API Key authentication
func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		authMode: "api_key",
		apiKey:   apiKey,
		baseURL:  baseURL,
		client:   &http.Client{},
	}
}

// NewClientWithJWT creates a new QWeather API client with JWT authentication
func NewClientWithJWT(privateKeyPath, keyID, projectID, baseURL string) (*Client, error) {
	// Read private key file
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// Parse PEM block
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ed25519Key, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not Ed25519")
	}

	logger.Info("QWeather JWT client initialized",
		zap.String("key_id", keyID),
		zap.String("project_id", projectID))

	return &Client{
		authMode:   "jwt",
		privateKey: ed25519Key,
		keyID:      keyID,
		projectID:  projectID,
		baseURL:    baseURL,
		client:     &http.Client{},
	}, nil
}

// base64URLEncode encodes bytes to base64url without padding
func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

// generateJWT creates a new JWT token using Ed25519 signature
func (c *Client) generateJWT() (string, error) {
	// Header
	header := map[string]string{
		"alg": "EdDSA",
		"kid": c.keyID,
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	// Payload
	now := time.Now().Unix()
	payload := map[string]interface{}{
		"sub": c.projectID,
		"iat": now - 30,       // 30 seconds before to account for clock skew
		"exp": now + 900 - 30, // 15 minutes validity
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Base64URL encode header and payload
	headerEncoded := base64URLEncode(headerJSON)
	payloadEncoded := base64URLEncode(payloadJSON)
	data := headerEncoded + "." + payloadEncoded

	// Sign with Ed25519
	signature := ed25519.Sign(c.privateKey, []byte(data))
	signatureEncoded := base64URLEncode(signature)

	// Combine to form JWT
	jwt := data + "." + signatureEncoded

	logger.Debug("JWT generated",
		zap.String("key_id", c.keyID),
		zap.Int64("iat", now-30),
		zap.Int64("exp", now+900-30))

	return jwt, nil
}

// doRequest sends HTTP request with proper authentication
func (c *Client) doRequest(requestURL string) (*http.Response, error) {
	// For api_key mode, append key to URL
	if c.authMode == "api_key" {
		if strings.Contains(requestURL, "?") {
			requestURL += "&key=" + c.apiKey
		} else {
			requestURL += "?key=" + c.apiKey
		}
	}

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header for JWT mode
	if c.authMode == "jwt" {
		token, err := c.generateJWT()
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return c.client.Do(req)
}

// GetLocationID retrieves the location ID for a city name
func (c *Client) GetLocationID(city string) (string, error) {
	logger.Debug("QWeather.GetLocationID called", zap.String("city", city))
	start := time.Now()

	params := url.Values{}
	params.Add("location", city)

	requestURL := fmt.Sprintf("%s/geo/v2/city/lookup?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to get location: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var geoResp GeoLocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return "", fmt.Errorf("failed to decode location response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", geoResp.Code),
		zap.Int("location_count", len(geoResp.Location)))

	if geoResp.Code != "200" || len(geoResp.Location) == 0 {
		logger.Warn("Location not found",
			zap.String("city", city),
			zap.String("api_code", geoResp.Code))
		return "", fmt.Errorf("location not found for city: %s", city)
	}

	logger.Debug("Location ID retrieved",
		zap.String("city", city),
		zap.String("location_id", geoResp.Location[0].ID),
		zap.Duration("duration", time.Since(start)))
	return geoResp.Location[0].ID, nil
}

// GetLocation retrieves the location details for a city name
func (c *Client) GetLocation(city string) (*GeoLocation, error) {
	logger.Debug("QWeather.GetLocation called", zap.String("city", city))
	start := time.Now()

	params := url.Values{}
	params.Add("location", city)

	requestURL := fmt.Sprintf("%s/geo/v2/city/lookup?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get location: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var geoResp GeoLocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode location response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", geoResp.Code),
		zap.Int("location_count", len(geoResp.Location)))

	if geoResp.Code != "200" || len(geoResp.Location) == 0 {
		logger.Warn("Location not found",
			zap.String("city", city),
			zap.String("api_code", geoResp.Code))
		return nil, fmt.Errorf("location not found for city: %s", city)
	}

	logger.Debug("Location retrieved",
		zap.String("city", city),
		zap.String("location_id", geoResp.Location[0].ID),
		zap.String("lat", geoResp.Location[0].Lat),
		zap.String("lon", geoResp.Location[0].Lon),
		zap.Duration("duration", time.Since(start)))
	return &geoResp.Location[0], nil
}

// GetCurrentWeather retrieves current weather for a location
func (c *Client) GetCurrentWeather(locationID string) (*CurrentWeather, error) {
	logger.Debug("QWeather.GetCurrentWeather called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)

	requestURL := fmt.Sprintf("%s/v7/weather/now?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get weather: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", weatherResp.Code))

	if weatherResp.Code != "200" {
		logger.Warn("Weather API error",
			zap.String("location_id", locationID),
			zap.String("api_code", weatherResp.Code))
		return nil, fmt.Errorf("weather API returned code: %s", weatherResp.Code)
	}

	logger.Debug("Current weather retrieved",
		zap.String("location_id", locationID),
		zap.String("temp", weatherResp.Now.Temp),
		zap.String("text", weatherResp.Now.Text),
		zap.Duration("duration", time.Since(start)))
	return &weatherResp.Now, nil
}

// GetLifeIndices retrieves life indices (clothing, UV, sports, etc.) for a location
func (c *Client) GetLifeIndices(locationID string) ([]LifeIndex, error) {
	logger.Debug("QWeather.GetLifeIndices called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)
	params.Add("type", "0") // 0 = all indices

	requestURL := fmt.Sprintf("%s/v7/indices/1d?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get life indices: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var indicesResp LifeIndicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&indicesResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode life indices response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", indicesResp.Code))

	if indicesResp.Code != "200" {
		logger.Warn("Life indices API error",
			zap.String("location_id", locationID),
			zap.String("api_code", indicesResp.Code))
		return nil, fmt.Errorf("life indices API returned code: %s", indicesResp.Code)
	}

	logger.Debug("Life indices retrieved",
		zap.String("location_id", locationID),
		zap.Int("indices_count", len(indicesResp.Daily)),
		zap.Duration("duration", time.Since(start)))
	return indicesResp.Daily, nil
}

// GetDailyForecast retrieves daily weather forecast for a location
func (c *Client) GetDailyForecast(locationID string) (*DailyForecast, error) {
	logger.Debug("QWeather.GetDailyForecast called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)

	requestURL := fmt.Sprintf("%s/v7/weather/3d?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get daily forecast: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var forecastResp DailyForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&forecastResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode daily forecast response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", forecastResp.Code))

	if forecastResp.Code != "200" || len(forecastResp.Daily) == 0 {
		logger.Warn("Daily forecast API error",
			zap.String("location_id", locationID),
			zap.String("api_code", forecastResp.Code))
		return nil, fmt.Errorf("daily forecast API returned code: %s", forecastResp.Code)
	}

	logger.Debug("Daily forecast retrieved",
		zap.String("location_id", locationID),
		zap.String("tempMax", forecastResp.Daily[0].TempMax),
		zap.String("tempMin", forecastResp.Daily[0].TempMin),
		zap.Duration("duration", time.Since(start)))
	return &forecastResp.Daily[0], nil
}

// GetAirQuality retrieves current air quality for a location
// Deprecated: Use GetAirQualityCurrent instead. This method uses the deprecated v7 API.
func (c *Client) GetAirQuality(locationID string) (*AirNow, error) {
	logger.Debug("QWeather.GetAirQuality called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)

	requestURL := fmt.Sprintf("%s/v7/air/now?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get air quality: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var airResp AirNowResponse
	if err := json.NewDecoder(resp.Body).Decode(&airResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode air quality response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", airResp.Code))

	if airResp.Code != "200" {
		logger.Warn("Air quality API error",
			zap.String("location_id", locationID),
			zap.String("api_code", airResp.Code))
		return nil, fmt.Errorf("air quality API returned code: %s", airResp.Code)
	}

	logger.Debug("Air quality retrieved",
		zap.String("location_id", locationID),
		zap.String("aqi", airResp.Now.Aqi),
		zap.String("category", airResp.Now.Category),
		zap.Duration("duration", time.Since(start)))
	return &airResp.Now, nil
}

// GetAirQualityCurrent retrieves current air quality using v1 API
func (c *Client) GetAirQualityCurrent(lat, lon string) (*AirQualityResponse, error) {
	logger.Debug("QWeather.GetAirQualityCurrent called", zap.String("lat", lat), zap.String("lon", lon))
	start := time.Now()

	// v1 API path: /airquality/v1/current/{lat}/{lon}
	// Note: The baseURL usually includes https://api.qweather.com or similar.
	// We need to construct the URL correctly.
	// Assuming baseURL is like "https://api.qweather.com/v7", we might need to adjust.
	// However, usually baseURL is just the host. Let's assume baseURL is the host root for now,
	// or we replace "/v7" if it's there.
	// Actually, standard QWeather baseURL is "https://dev.qweather.com" or "https://api.qweather.com".
	// The v7 endpoints are like /v7/weather/now.
	// The v1 endpoint is /airquality/v1/current/...
	// So we just append /airquality/v1/current/... to the base URL.

	requestURL := fmt.Sprintf("%s/airquality/v1/current/%s/%s", c.baseURL, lat, lon)

	// Add query parameters (lang, etc. if needed, but currently none required)
	params := url.Values{}
	// params.Add("lang", "zh") // Optional
	if len(params) > 0 {
		requestURL += "?" + params.Encode()
	}

	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get air quality: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var airResp AirQualityResponse
	if err := json.NewDecoder(resp.Body).Decode(&airResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode air quality response: %w", err)
	}

	// Check if response is valid (v1 might not have "code" field in root like v7)
	// Based on docs, it returns JSON directly.
	// We should check if Indexes is empty or if there's an error field (not standard in success response).
	// Let's assume if we decoded it and got data, it's fine.

	logger.Debug("Air quality retrieved",
		zap.String("lat", lat),
		zap.String("lon", lon),
		zap.Int("indexes_count", len(airResp.Indexes)),
		zap.Duration("duration", time.Since(start)))
	return &airResp, nil
}

// GetAirDailyForecast retrieves daily air quality forecast for a location
func (c *Client) GetAirDailyForecast(locationID string) ([]AirDaily, error) {
	logger.Debug("QWeather.GetAirDailyForecast called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)

	requestURL := fmt.Sprintf("%s/v7/air/5d?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get air daily forecast: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var airResp AirDailyResponse
	if err := json.NewDecoder(resp.Body).Decode(&airResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode air daily forecast response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", airResp.Code))

	if airResp.Code != "200" {
		logger.Warn("Air daily forecast API error",
			zap.String("location_id", locationID),
			zap.String("api_code", airResp.Code))
		return nil, fmt.Errorf("air daily forecast API returned code: %s", airResp.Code)
	}

	logger.Debug("Air daily forecast retrieved",
		zap.String("location_id", locationID),
		zap.Int("forecast_count", len(airResp.Daily)),
		zap.Duration("duration", time.Since(start)))
	return airResp.Daily, nil
}

// GetWarning retrieves weather warnings for a location
func (c *Client) GetWarning(locationID string) ([]Warning, error) {
	logger.Debug("QWeather.GetWarning called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)

	requestURL := fmt.Sprintf("%s/v7/warning/now?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get weather warning: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var warningResp WarningResponse
	if err := json.NewDecoder(resp.Body).Decode(&warningResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode warning response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", warningResp.Code))

	if warningResp.Code != "200" {
		logger.Warn("Warning API error",
			zap.String("location_id", locationID),
			zap.String("api_code", warningResp.Code))
		return nil, fmt.Errorf("warning API returned code: %s", warningResp.Code)
	}

	logger.Debug("Weather warnings retrieved",
		zap.String("location_id", locationID),
		zap.Int("warning_count", len(warningResp.Warning)),
		zap.Duration("duration", time.Since(start)))
	return warningResp.Warning, nil
}
