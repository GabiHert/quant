package usecase

// SendNotification defines the interface for sending system notifications.
type SendNotification interface {
	SendNotification(title, message string) error
}
