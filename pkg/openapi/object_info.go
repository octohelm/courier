package openapi

// https://spec.openapis.org/oas/latest.html#infoObject
type InfoObject struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitzero"`
	TermsOfService string   `json:"termsOfService,omitzero"`
	Contact        *Contact `json:"contact,omitzero"`
	License        *License `json:"license,omitzero"`
	Version        string   `json:"version"`
}

type Contact struct {
	Name  string `json:"name,omitzero"`
	URL   string `json:"url,omitzero"`
	Email string `json:"email,omitzero"`
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitzero"`
}
