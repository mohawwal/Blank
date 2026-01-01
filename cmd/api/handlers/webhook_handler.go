package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *Handler) WhatsAppWebhook(c echo.Context) error {
	from := c.FormValue("From")
	body := c.FormValue("Body")

	if from == "" || body == "" {
		return c.String(http.StatusBadRequest, "Missing From or Body parameter")
	}

	userPhone := strings.Replace(from, "whatsapp:", "", 1)
	cleanBody := strings.TrimSpace(body)
	profileName := c.FormValue("ProfileName")

	log.Printf("Received message from %s: %s", userPhone, cleanBody)

	// Process message asynchronously using the message processor
	go h.MessageProcessor.ProcessMessage(userPhone, profileName, cleanBody)

	return c.String(http.StatusOK, "Ok")
}
