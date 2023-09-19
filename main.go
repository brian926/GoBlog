package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB
var tokenAuth *jwtauth.JWTAuth

const Secret = "jwt-secret"

type Article struct {
	ID      int           `json:"id"`
	Title   string        `json:"title"`
	Content template.HTML `json:"content"`
}

func catch(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func MakeToken(name string) string {
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"username": name})
	return tokenString
}

func init() {
	load := godotenv.Load()
	if load != nil {
		fmt.Print("Error loading .env file")
	}

	tokenAuth = jwtauth.New("HS256", []byte(Secret), nil)

	var err error
	db, err = connect()
	catch(err)
}

func router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(ChangeMethod)

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))

		r.Post("/upload", UploadHandler)
		r.Route("/articles", func(r chi.Router) {
			r.Get("/", NewArticle)
			r.Post("/", CreateArticle)
			r.Route("/{articleID}", func(r chi.Router) {
				r.Use(ArticleCtx)
				r.Put("/", UpdateArticle)    // PUT /articles/1234
				r.Delete("/", DeleteArticle) // DELETE /articles/1234
				r.Get("/edit", EditArticle)  // GET /articles/1234/edit
			})
		})
		r.Post("/signin", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			userName := r.PostForm.Get("username")
			userPassword := r.PostForm.Get("password")

			if userName == "" || userPassword == "" {
				http.Error(w, "missing user or password", http.StatusBadRequest)
				return
			}

			token := MakeToken(userName)

			http.SetCookie(w, &http.Cookie{
				HttpOnly: true,
				Expires:  time.Now().Add(7 * 24 * time.Hour),
				SameSite: http.SameSiteLaxMode,
				// Uncomment below for HTTPS:
				// Secure: true,
				Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
				Value: token,
			})
		})
	})

	// Public Routes
	r.Get("/", GetAllArticles)
	r.Get("/images/*", ServeImages) // Add this
	r.Route("/articles", func(r chi.Router) {
		r.Route("/{articleID}", func(r chi.Router) {
			r.Use(ArticleCtx)
			r.Get("/", GetArticle) // GET /articles/1234
		})
	})

	// r.Post("/signin", Signin)
	// r.Get("/welcome", Welcome)
	// r.Post("/refresh", Refresh)
	// r.Post("/logout", Logout)

	return r
}

func main() {
	port := "8080"

	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	log.Printf("Starting up on http://localhost:%s", port)

	log.Fatal(http.ListenAndServe(":"+port, router()))
}
