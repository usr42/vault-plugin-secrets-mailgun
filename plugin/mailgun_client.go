package mgsecret

import "github.com/mailgun/mailgun-go"

type MailgunClient interface {
	IsDomainValid() bool
	IsApiKeyValid() bool
	DeleteCredential(username string) error
	CreateCredential(login, password string) error
}

type mailgunClientImpl struct {
	*mailgun.MailgunImpl
}

func (client mailgunClientImpl) IsApiKeyValid() bool {
	if _, _, err := client.GetDomains(1, 0); err != nil {
		return false
	}
	return true
}

func (client mailgunClientImpl) IsDomainValid() bool {
	if _, _, err := client.GetCredentials(1, 0); err != nil {
		return false
	}
	return true
}

func DefaultMailgunClientFactory(domain, apiKey string) MailgunClient {
	return mailgunClientImpl{mailgun.NewMailgun(domain, apiKey)}
}
