package main

import "github.com/golang-jwt/jwt/v5"

type User struct {
	ID           int    `json:"-"`
	Username     string `json:"username" pg:"username"`
	PasswordHash string `json:"password" pg:"passwordhash"`
}

type UserRequest struct {
	Bookname string `json:"bookname" pg:"bookname"`
	Status   string `json:"status" pg:"status"`
}

type Items struct {
	ID       int    `json:"-"`
	UserID   int    `pg:"user_id"`
	Bookname string `pg:"bookname"`
	Status   string `json:"status" pg:"status"`
}

var uuid = struct{}{}

type JWTClaims struct {
	Username string
	UserId   string
	jwt.RegisteredClaims
}
