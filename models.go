package main

type LoginParameters struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type RegisterParameters struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}
