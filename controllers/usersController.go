package controllers

import (
	"auth/initializers"
	"auth/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context) {
	//getting the sign up information
	var body struct {
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("Error binding request data: %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	//hashing the password
	hashedPassowrd, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		log.Printf("Error hashing the password: %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed hashing the password",
		})
	}
	//preparing the information of the user
	user := models.User{Name: body.Name, Password: string(hashedPassowrd), Role: "User"}
	//entering the information into the user data base
	result := initializers.DB.Create(&user)
	if result.Error != nil { //if it failed to create the new user we get an error
		log.Printf("Error creating a new user: %v", result.Error)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Failed to creater new user",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func Login(c *gin.Context) {
	var body struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("Error binding request data: %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read JSON body",
		})
		return
	}
	var user models.User
	//Looking for the user with the entered name
	initializers.DB.Where("name = ?", body.Name).First(&user)
	if user.ID == 0 {
		log.Printf("Error finding the user")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "This user does not exist, please sign up",
		})
		return
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		log.Printf("Wrong password entered: %v", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// If we got here, it means we logged in with valid information.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Role": user.Role,
		"Sub":  user.ID,
		"Exp":  time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret.
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRETKEY")))
	if err != nil {
		log.Printf("Failed getting the encoded token: %v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create token",
		})
		return
	}
	// Sending the token back with a cookie.
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Auth", tokenString, 3600*24*30, "", "", false, true)

	// Respond with success status and the token.
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}

func Logout(c *gin.Context) {
	// Clear the authentication token by setting an expired cookie.
	c.SetCookie("Auth", "", -1, "", "", false, true)

	// Redirect to the login page.
	c.Redirect(http.StatusSeeOther, "/login")

	// Send a JSON response
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})

}
