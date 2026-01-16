package main

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

type identityToken struct {
	Exp     int     `json:"exp"`
	Iat     int     `json:"iat"`
	Iss     string  `json:"iss"`
	Jti     string  `json:"jti"`
	Scope   string  `json:"scope"`
	Sub     string  `json:"sub"`
	Profile profileInfo `json:"profile"`
}


type sessionToken struct {
	Exp   int    `json:"exp"`
	Iat   int    `json:"iat"`
	Iss   string `json:"iss"`
	Jti   string `json:"jti"`
	Scope string `json:"scope"`
	Sub   string `json:"sub"`
}

type profileInfo struct {
	Username     string   `json:"username"`
	Entitlements []string `json:"entitlements"`
	Skin         string   `json:"skin"`
}
