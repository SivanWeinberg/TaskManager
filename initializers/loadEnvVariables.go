package initializers

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal("Error loading .env")
	}
}
