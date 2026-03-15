package usecase

// SetLoginItem defines the interface for enabling or disabling start-on-login behavior.
type SetLoginItem interface {
	SetLoginItem(enabled bool) error
}
