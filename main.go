package main

import (
	"auth/initializers"
	"auth/routes"
	"html/template"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables() //loading environment variables
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}
func main() {
	website := gin.Default()
	tmpl := template.Must(template.ParseGlob("templates/*.html")) //Getting all the html templates
	routes.RunRoutes(website, tmpl)

	website.Run(":3000") //Running the server

}
