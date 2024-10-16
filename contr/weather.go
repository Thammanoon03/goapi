package contr

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const apikey = "bbbdf73de569ef654e83c6e2c06f4ed0"
const city = "Sisaket"
const apiURL = "http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric"

type WeatherResponse struct {
	Name string `json:"name"`
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"` // Fixed typo here
	} `json:"weather"`
}

func GetWeather(c *gin.Context) {
	url := fmt.Sprintf(apiURL, city, apikey)
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch weather data"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch weather data"})
		return
	}

	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding weather data"})
		return
	}

	c.JSON(http.StatusOK, weather)
}
