package main

import (
	"html/template"
	"net/http"

	"github.com/go-chi/jwtauth"
)

var tokenAuth *jwtauth.JWTAuth

const Secret = "jwt-secret"

type User struct {
	Username string
}

type PageData struct {
	User *User
}

func MakeToken(name string) string {
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"username": name})
	return tokenString
}

func ParseTemplates(r *http.Request, templates []string) (*template.Template, *PageData) {
	_, claims, _ := jwtauth.FromContext(r.Context())

	tmpl := template.Must(template.ParseFiles(templates...))

	data := &PageData{
		User: nil,
	}

	if claims["username"] != nil {
		data.User = &User{
			Username: claims["username"].(string),
		}
	}

	return tmpl, data
}
