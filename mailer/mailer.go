package mailer

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/netlify/gotrue/conf"
	"github.com/netlify/gotrue/models"
)

func getEnvBool(key string) bool{
	v := os.Getenv(key)
	if len(v) == 0 {
		return false
	}

	if strings.ToLower(v) != "true" {
		return false
	}

	return true
}

var withoutFragment = getEnvBool("GOTRUE_WITHOUT_FRAGMENT")
var addURLType = getEnvBool("GOTRUE_ADD_URL_TYPE")

// Mailer defines the interface a mailer must implement.
type Mailer interface {
	Send(user *models.User, subject, body string, data map[string]interface{}) error
	InviteMail(user *models.User, referrerURL string) error
	ConfirmationMail(user *models.User, referrerURL string) error
	RecoveryMail(user *models.User, referrerURL string) error
	EmailChangeMail(user *models.User, referrerURL string) error
	ValidateEmail(email string) error
}

// NewMailer returns a new gotrue mailer
func NewMailer(instanceConfig *conf.Configuration) Mailer {
	if instanceConfig.SMTP.Host == "" {
		return &noopMailer{}
	}

	return &TemplateMailer{
		SiteURL: instanceConfig.SiteURL,
		Config:  instanceConfig,
		Mailer: &MailmeMailer{
			Host:    instanceConfig.SMTP.Host,
			Port:    instanceConfig.SMTP.Port,
			User:    instanceConfig.SMTP.User,
			Pass:    instanceConfig.SMTP.Pass,
			From:    instanceConfig.SMTP.AdminEmail,
			BaseURL: instanceConfig.SiteURL,
		},
	}
}

func withDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func getSiteURL(referrerURL, siteURL, filepath, fragment string, uType string) (string, error) {
	baseURL := siteURL
	if filepath == "" && referrerURL != "" {
		baseURL = referrerURL
	}

	site, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	if filepath != "" {
		path, err := url.Parse(filepath)
		if err != nil {
			return "", err
		}
		site = site.ResolveReference(path)
	}

	site.Fragment = fragment
	out := site.String()
	if withoutFragment {
		if !addURLType {
			out = strings.ReplaceAll(out, "#", "?")
		} else {
			out = strings.ReplaceAll(out, "#", fmt.Sprintf("%s/?", uType))
		}
	}

	return out, nil
}

var urlRegexp = regexp.MustCompile(`^https?://[^/]+`)

func enforceRelativeURL(url string) string {
	return urlRegexp.ReplaceAllString(url, "")
}
