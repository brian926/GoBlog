package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var router *chi.Mux
var db *sql.DB

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

func init() {
	load := godotenv.Load()
	if load != nil {
		fmt.Print("Error loading .env file")
	}

	router = chi.NewRouter()
	router.Use(middleware.Recoverer)

	var err error
	db, err = connect()
	catch(err)
}

func main() {
	router = chi.NewRouter()
	router.Use(middleware.Recoverer)

	var err error
	db, err = connect()
	catch(err)

	router.Use(ChangeMethod)
	router.Get("/", GetAllArticles)
	router.Post("/upload", UploadHandler) // Add this
	router.Get("/images/*", ServeImages)  // Add this
	router.Route("/articles", func(r chi.Router) {
		r.Get("/", NewArticle)
		r.Post("/", CreateArticle)
		r.Route("/{articleID}", func(r chi.Router) {
			r.Use(ArticleCtx)
			r.Get("/", GetArticle)       // GET /articles/1234
			r.Put("/", UpdateArticle)    // PUT /articles/1234
			r.Delete("/", DeleteArticle) // DELETE /articles/1234
			r.Get("/edit", EditArticle)  // GET /articles/1234/edit
		})
	})
	router.Post("/signin", Signin)
	router.Get("/welcome", Welcome)
	router.Post("/refresh", Refresh)
	router.Post("/logout", Logout)

	err = http.ListenAndServe(":8005", router)
	catch(err)
}
