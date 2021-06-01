package main

type LoginParameters struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type RegisterParameters struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginResult struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	JWTToken string `json:"jwt_token"`
}

func NewLoginResult(id int, name, jwtToken string) *LoginResult {
	return &LoginResult{
		ID:       id,
		Name:     name,
		JWTToken: jwtToken,
	}
}

type JoinParameters struct {
	IsPublisher bool `json:"isPublisher"`
}
