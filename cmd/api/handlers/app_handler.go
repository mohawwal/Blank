package handlers

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HealthCheck(c echo.Context) error {
	healthCheckStruct := struct {
		Health bool
	}{
		Health: true,
	}
	return c.JSON(http.StatusOK, healthCheckStruct)
}
