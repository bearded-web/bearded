package token

type TokenEntity struct {
	Name string `json:"name,omitempty" description:"token name" validate:"max=256"`
	//	Scopes  []string      `json:"scopes,omitempty"`
}
