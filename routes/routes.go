package routes

import (
	"authentication/controllers"
	"authentication/middleware"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RunRoutes(website *gin.Engine, tmpl *template.Template) {
	website.POST("/", controllers.SignUp)
	website.POST("/login", controllers.Login)
	website.POST("/createTask", middleware.RequiredAuth, controllers.CreateTask)
	website.DELETE("/deleteTask", middleware.RequiredAuth, controllers.DeleteTask)
	website.DELETE("/deleteTask/:title", middleware.RequiredAuth, controllers.DeleteTask)
	website.PUT("/updateTask", middleware.RequiredAuth, controllers.UpdateTask)
	website.GET("/readTasks", middleware.RequiredAuth, controllers.ReadAllTasks)
	website.GET("/logout", controllers.Logout)
	website.LoadHTMLFiles("templates/index.html", "templates/login.html", "templates/createTask.html", "templates/deleteTask.html", "templates/updateTask.html", "templates/readTasks.html",
		"templates/unAuthorized.html")
	website.Static("/static", "static")

	// Define a route to render the signup form
	website.GET("/", func(c *gin.Context) {
		renderTemplate(c, tmpl, "index.html", nil)
	})
	website.GET("/login", func(c *gin.Context) {
		renderTemplate(c, tmpl, "login.html", nil)
	})
	website.GET("/createTask", middleware.RequiredAuth, func(c *gin.Context) {
		renderTemplate(c, tmpl, "createTask.html", nil)
	})
	website.GET("/deleteTask/:title", middleware.RequiredAuth, controllers.DeleteTask)
	website.GET("/deleteTask", middleware.RequiredAuth, func(c *gin.Context) {
		renderTemplate(c, tmpl, "deleteTask.html", nil)
	})
	website.GET("/updateTask", middleware.RequiredAuth, func(c *gin.Context) {
		renderTemplate(c, tmpl, "updateTask.html", nil)
	})
}
func renderTemplate(c *gin.Context, tmpl *template.Template, name string, data interface{}) {
	err := tmpl.ExecuteTemplate(c.Writer, name, data)
	if err != nil {
		fmt.Println("An error occurred:", err.Error())
		c.HTML(http.StatusInternalServerError, "unAuthorized.html", gin.H{"error": err.Error()})
	}
}
