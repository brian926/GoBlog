package main

import (
	"database/sql"
	"fmt"
	"os"
)

func connect() (*sql.DB, error) {
	dbinfo := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

	var err error
	db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, err
	}

	sqlStmt := `
	create table if not exists articles (id serial PRIMARY KEY, title text, content text);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func dbCreateArticle(article *Article) error {
	query, err := db.Prepare("insert into articles(title,content) values ($1,$2);")

	if err != nil {
		return err
	}
	_, err = query.Exec(article.Title, article.Content)
	defer query.Close()
	if err != nil {
		return err
	}

	return nil
}

func dbGetAllArticles() ([]*Article, error) {
	query, err := db.Prepare("select id, title, content from articles")
	if err != nil {
		return nil, err
	}

	result, err := query.Query()
	defer query.Close()
	if err != nil {
		return nil, err
	}
	articles := make([]*Article, 0)
	for result.Next() {
		data := new(Article)
		err := result.Scan(
			&data.ID,
			&data.Title,
			&data.Content,
		)
		if err != nil {
			return nil, err
		}
		articles = append(articles, data)
	}

	return articles, nil
}

func dbGetArticle(articleID string) (*Article, error) {
	query, err := db.Prepare("select id, title, content from articles where id = $1")
	if err != nil {
		return nil, err
	}

	result := query.QueryRow(articleID)
	defer query.Close()
	data := new(Article)
	err = result.Scan(&data.ID, &data.Title, &data.Content)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func dbUpdateArticle(id string, article *Article) error {
	query, err := db.Prepare("update articles set (title, content) = ($1,$2) where id=$3")
	if err != nil {
		return err
	}
	_, err = query.Exec(article.Title, article.Content, id)
	defer query.Close()
	if err != nil {
		return err
	}

	return nil
}

func dbDeleteArticle(id string) error {
	query, err := db.Prepare("delete from articles where id=$1")
	if err != nil {
		return err
	}
	_, err = query.Exec(id)
	defer query.Close()
	if err != nil {
		return err
	}

	return nil
}
