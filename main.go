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
	"github.com/lestrrat-go/jwx/jwt"
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

type User struct {
	Username string
}

type PageData struct {
	User *User
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

func LoggedInRedirector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, _ := jwtauth.FromContext(r.Context())

		if token != nil && jwt.Validate(token) == nil {
			http.Redirect(w, r, "/", http.StatusFound)
		}

		next.ServeHTTP(w, r)
	})
}

func UnloggedInRedirector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, _ := jwtauth.FromContext(r.Context())

		if token == nil || jwt.Validate(token) != nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
		}

		next.ServeHTTP(w, r)
	})
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

		r.Use(UnloggedInRedirector)

		r.Post("/upload", UploadHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))

		r.Use(LoggedInRedirector)

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

			http.Redirect(w, r, "/", http.StatusSeeOther)
		})
	})

	// Public Route
	r.Get("/images/*", ServeImages) // Add this
	r.Get("/", GetAllArticles)

	r.Route("/articles", func(r chi.Router) {
		// Private
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))

			r.Use(UnloggedInRedirector)
			r.Get("/", NewArticle)
			r.Post("/", CreateArticle)
		})
		r.Route("/{articleID}", func(r chi.Router) {
			r.Use(ArticleCtx)
			// Public Route
			r.Get("/", GetArticle) // GET /articles/1234
			// Private Routes
			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(tokenAuth))

				r.Use(UnloggedInRedirector)

				r.Put("/", UpdateArticle)    // PUT /articles/1234
				r.Delete("/", DeleteArticle) // DELETE /articles/1234
				r.Get("/edit", EditArticle)  // GET /articles/1234/edit
			})
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))

		r.Use(LoggedInRedirector)
		r.Get("/signin", func(w http.ResponseWriter, r *http.Request) {
			tmpl, data := ParseTemplates(r, []string{"partials/navbar.html", "templates/signin.html"})

			tmpl.ExecuteTemplate(w, "signin", data)
		})
	})
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
