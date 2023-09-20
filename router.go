package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwt"
)

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
