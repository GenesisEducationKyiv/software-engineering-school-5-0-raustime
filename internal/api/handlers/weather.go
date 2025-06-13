package handlers

import (
	"errors"
	"net/http"

	"weatherapi/internal/openweatherapi"

	"github.com/gin-gonic/gin"
)

type WeatherResponse struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
}

func GetWeatherHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		if city == "" {
			c.Status(http.StatusBadRequest) // 400 без тіла
			return
		}

		weather, err := openweatherapi.FetchWeather(city)
		if err != nil {
			if errors.Is(err, openweatherapi.ErrCityNotFound) {
				c.Status(404)
				return
			}
			c.Status(400)
			return
		}

		c.JSON(http.StatusOK, WeatherResponse{
			Temperature: weather.Temperature,
			Humidity:    weather.Humidity,
			Description: weather.Description,
		})
	}
}
