package valueobject

type IntegrationType string

const (
	IntegrationTypeOAuth  IntegrationType = "OAUTH"
	IntegrationTypeManual IntegrationType = "MANUAL"
)

func (i IntegrationType) String() string {
	return string(i)
}

func (i IntegrationType) IsValid() bool {
	switch i {
	case IntegrationTypeOAuth, IntegrationTypeManual:
		return true
	default:
		return false
	}
}
