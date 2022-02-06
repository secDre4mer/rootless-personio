package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"goPersonio/pkg/personio"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	personioUser := os.Getenv("PERSONIO_USER")
	personioPassword := os.Getenv("PERSONIO_PASSWORD")
	personioBaseURL := os.Getenv("PERSONIO_BASE_URL")

	p := personio.NewPersonio(personioBaseURL, personioUser, personioPassword)
	p.LoginToPersonio()
	p.SetWorkingTimes(time.Now(), time.Now().Add(time.Minute*4))
}
