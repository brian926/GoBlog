package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/jwtauth"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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

	tokenAuth = jwtauth.New("HS256", []byte(Secret), nil)

	var err error
	db, err = connect()
	catch(err)
}

func main() {
	port := "8080"

	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	log.Printf("Starting up on http://localhost:%s", port)

	log.Fatal(http.ListenAndServe(":"+port, router()))
}
