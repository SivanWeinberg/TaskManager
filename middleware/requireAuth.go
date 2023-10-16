package middleware

import (
	"auth/initializers"
	"auth/models"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func RequiredAuth(c *gin.Context) {
	cookieToken, err := c.Cookie("Auth") //getting the Cookie named Auth information
	if err != nil {                      // if error is not nil it shows the user unAuthorized.html
		c.HTML(http.StatusUnauthorized, "unAuthorized.html", gin.H{"message": "You do not have permission to access this page. Please log in."})
		c.Abort()
		return
	}

	token, _ := jwt.Parse(cookieToken, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("SECRETKEY")), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		if float64(time.Now().Unix()) > claims["Exp"].(float64) { //if the cookie expired the person won't be able to enter the user parts of the website
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		var user models.User
		initializers.DB.First(&user, claims["Sub"])

		if user.ID == 0 { // if it failed getting the user, we want to show an error
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if user.Role != "User" { //checking if the user we got has the User role
			c.HTML(http.StatusUnauthorized, "unAuthorized.html", gin.H{"message": "You do not have permission to access this page. Please log in."})
			c.Abort()
		}
		c.Set("user", user)
		c.Next() // if everything passed we want to move to the next handler
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
}
