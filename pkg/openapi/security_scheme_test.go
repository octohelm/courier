package openapi

import (
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/onsi/gomega"
)

func TestSecurityScheme(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		actual := SecurityScheme{}

		expect := `{"type":""}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expect))
	})

	t.Run("with api key", func(t *testing.T) {
		actual := NewAPIKeySecurityScheme("AccessToken", PositionHeader)

		expect := `{"type":"apiKey","name":"AccessToken","in":"header"}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expect))
	})

	t.Run("with http basic", func(t *testing.T) {
		actual := NewHTTPSecurityScheme("basic", "")

		expect := `{"type":"http","scheme":"basic"}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expect))
	})

	t.Run("with http bearer", func(t *testing.T) {
		actual := NewHTTPSecurityScheme("bearer", "JWT")

		expect := `{"type":"http","scheme":"bearer","bearerFormat":"JWT"}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expect))
	})

	t.Run("with open id connect", func(t *testing.T) {
		actual := NewOpenIdConnectSecurityScheme("http://xx.com")

		expect := `{"type":"openIdConnect","openIdConnectUrl":"http://xx.com"}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expect))
	})

	t.Run("with open id connect", func(t *testing.T) {
		actual := NewOAuth2SecurityScheme(OAuthFlowsObject{
			Implicit: NewOAuthFlow(
				"https://example.com/api/oauth",
				"https://example.com/api/oauth/token_token",
				"https://example.com/api/oauth/reflesh_token",
				map[string]string{
					"write:pets": "modify pets in your account",
					"read:pets":  "read your pets",
				},
			),
		})

		expect := `{"type":"oauth2","flows":{"implicit":{"authorizationUrl":"https://example.com/api/oauth","tokenUrl":"https://example.com/api/oauth/token_token","refreshUrl":"https://example.com/api/oauth/reflesh_token","scopes":{"read:pets":"read your pets","write:pets":"modify pets in your account"}}}}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expect))
	})
}

func TestSecurityRequirement(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(SecurityRequirement{})).To(gomega.Equal(`{}`))
	})

	t.Run("Non-OAuth2 Security Requirement", func(t *testing.T) {
		actual := SecurityRequirement{"api_key": []string{}}

		expect := `{"api_key":[]}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expect))
	})

	t.Run("OAuth2 Security Requirement", func(t *testing.T) {
		actual := SecurityRequirement{"petstore_auth": []string{
			"write:pets",
			"read:pets",
		}}

		expect := `{"petstore_auth":["write:pets","read:pets"]}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expect))
	})

}
