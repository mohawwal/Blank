package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// PaymentRedirect redirects to the full Paystack payment URL
func (h *Handler) PaymentRedirect(c echo.Context) error {
	// Get the reference parameter from query string
	ref := c.QueryParam("ref")

	if ref == "" {
		return c.String(http.StatusBadRequest, "Missing payment reference")
	}

	// Redirect to the full Paystack URL
	return c.Redirect(http.StatusTemporaryRedirect, ref)
}
