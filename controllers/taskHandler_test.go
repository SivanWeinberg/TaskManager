package controllers

import (
	"authentication/initializers"
	"authentication/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateTask(t *testing.T) {
	//loading database
	initializers.LoadEnvVariables() //loading environment variables
	initializers.ConnectToDb()
	initializers.SyncDatabase()
	initializers.DB.Unscoped().Where("user_id = ?", 0).Delete(&models.Task{})

	router := gin.Default()
	user := models.User{
		Name:     "TestUser",
		Password: "11111",
		Role:     "User",
	}
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.POST("/createTask", CreateTask)

	//tasks to test on the create function
	cases := []struct {
		title, desc            string
		day, month, year, code int
	}{
		{"Shake", "cook a lot", 1, 1, 1345, 200},
		{"", "", 1, 2, 1934, 400},
		{"bluh", "", 1, 2, 1934, 400},
		{"Dance", "Party", -1, 2, 1934, 400},
	}
	for _, c := range cases {
		//json body for the post request.
		bodyString := `{"title":"` + c.title + `","description":"` + c.desc + `","day":` + strconv.Itoa(c.day) + `,"month":` + strconv.Itoa(c.month) + `,"year":` + strconv.Itoa(c.year) + `}`
		body := strings.NewReader(bodyString)
		req, _ := http.NewRequest("POST", "/createTask", body)
		req.Header.Add("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		//if it won't return the correct code the fail will fail
		if resp.Code != c.code {
			t.Errorf("Expected status code %d, but got %d", c.code, resp.Code)
		}

	}
}

func TestUpdateTask(t *testing.T) {
	// Your initialization code here...

	router := gin.Default()

	user := models.User{
		Name:     "TestUser",
		Password: "11111",
		Role:     "User",
	}
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.PUT("/updateTask", UpdateTask)

	cases := []struct {
		currentTitle   string
		currentDesc    string
		currentDueDate string
		userId         int
		titleToUpdate  string
		newTitle       string
		newDescription string
		newDay         string
		newMonth       string
		newYear        string
		code           int
	}{
		{"Shake", "shake it o", "11/12/1900", 0, "Shake", "Cook", "cook a lot", "1", "1", "1345", 200},
		{"Shake", "shake it o", "11/12/1900", 0, "Dance", "Cook", "cook a lot", "1", "1", "1345", 400},
		{"Run", "shake it o", "11/12/1900", 0, "Run", "Party", "", "", "", "", 200},
		{"Run", "shake it o", "11/12/1900", 0, "Run", "Party", "", "-1", "", "", 500},
		{"Run", "shake it o", "11/12/1900", 0, "Run", "Party", "", "", "-1", "", 500},
	}

	for _, c := range cases {
		//creating a task to run a test on it
		task := models.Task{
			Title:       c.currentTitle,
			Description: c.currentDesc,
			DueDate:     c.currentDueDate,
			UserId:      uint(c.userId), // You may need to cast userId to uint if it's not already.

			// Other fields you want to set
		}
		initializers.DB.Create(&task)
		//request body
		updateData := struct {
			CurrentTitle   string `json:"titleToUpdate"`
			NewTitle       string `json:"newTitle"`
			NewDescription string `json:"newDescription"`
			NewDay         string `json:"newDay"`
			NewMonth       string `json:"newMonth"`
			NewYear        string `json:"newYear"`
		}{
			CurrentTitle:   c.titleToUpdate,
			NewTitle:       c.newTitle,
			NewDescription: c.newDescription,
			NewDay:         c.newDay,
			NewMonth:       c.newMonth,
			NewYear:        c.newYear,
		}
		// Marshal the struct into a JSON string
		body, _ := json.Marshal(updateData)

		// Create the request with the JSON body
		req, _ := http.NewRequest("PUT", "/updateTask", bytes.NewReader(body))
		req.Header.Add("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req) // making the request and getting the respond to resp
		initializers.DB.Unscoped().Where("user_id = ?", 0).Delete(&models.Task{})
		//if we didn't recieve the expected code the test will fail
		if resp.Code != c.code {
			t.Errorf("Expected status code %d, but got %d", c.code, resp.Code)
		}
	}
}

func TestDeleteTask(t *testing.T) {
	initializers.LoadEnvVariables() //loading environment variables
	initializers.ConnectToDb()
	initializers.SyncDatabase()

	router := gin.Default()
	user := models.User{
		Name:     "TestUser",
		Password: "11111",
		Role:     "User",
	}
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.DELETE("/deleteTask", DeleteTask)

	cases := []struct {
		currentTitle   string
		currentDesc    string
		currentDueDate string
		userId         int
		titleToDelete  string
		code           int
	}{
		{"Shake", "cook a lot", "11/09/1923", 0, "Shake", 200},
		{"Shake", "cook a lot", "11/09/1923", 0, "", 400},
		{"Shake", "cook a lot", "11/09/1923", 0, "Run", 404},
	}
	for _, c := range cases {
		task := models.Task{
			Title:       c.currentTitle,
			Description: c.currentDesc,
			DueDate:     c.currentDueDate,
			UserId:      uint(c.userId), // You may need to cast userId to uint if it's not already.

			// Other fields you want to set
		}
		initializers.DB.Create(&task)

		bodyString := `{
			"titleToDelete": "` + c.titleToDelete + `"
		}`
		body := strings.NewReader(bodyString)
		fmt.Println(body)
		req, _ := http.NewRequest("DELETE", "/deleteTask", body)
		req.Header.Add("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		//if it didn't delete the task we created we want to delete it from the database
		if resp.Code != 200 {
			initializers.DB.Unscoped().Where("title = ? AND user_id = ?", c.currentTitle, c.userId).Delete(&models.Task{})

		}
		if resp.Code != c.code {
			t.Errorf("Expected status code %d, but got %d", c.code, resp.Code)
		}

	}
}
