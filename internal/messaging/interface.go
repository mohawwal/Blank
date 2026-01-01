package messaging

// MessagingService defines the interface for sending messages
type MessagingService interface {
	SendMessage(to string, body string) error
	SendTemplateMessage(to string, templateData TemplateData) error
}

// TemplateData holds data for template messages
type TemplateData struct {
	TemplateName string
	Variables    map[string]string
}
