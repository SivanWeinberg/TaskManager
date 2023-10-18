package controllers

import (
	"autentication/models"
	"authentication/initializers"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateTask(c *gin.Context) {
	//getting task information from the user
	var taskBody struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		Day         int    `json:"day" binding:"required"`
		Month       int    `json:"month" binding:"required"`
		Year        int    `json:"year" binding:"required"`
	}

	if err := c.ShouldBindJSON(&taskBody); err != nil {
		log.Printf("Error binding request data: %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read JSON body",
		})
		return
	}
	//Getting the user information that was saved in middleware
	user, _ := c.Get("user")
	fmt.Println(user)
	userObj := user.(models.User)
	//Checking if the date entered is valid
	dueDate := validateDate(taskBody.Day, taskBody.Month, taskBody.Year, c)

	if dueDate == "" { //If the dueDate is an empty string it means the date is invalid.
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid date.",
		})
		return
	}
	task := models.Task{Title: taskBody.Title, Description: taskBody.Description, DueDate: dueDate, UserId: userObj.ID} //Creating the task we will enter to the data base
	createTask := initializers.DB.Create(&task)                                                                         //Creating the new task

	if createTask.Error != nil {
		log.Printf("Error creating a task: %v", createTask.Error)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Failed to creater new task",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "created the task"}) //If we passed everything until now we want to announce that the task was created

}

func DeleteTask(c *gin.Context) {
	user, _ := c.Get("user")
	userObj := user.(models.User)
	getUserId := userObj.ID
	if userObj.Role != "User" { // Checking if the user got a permission
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "you need to log in to get permission",
		})
		return
	}
	//If the url contains a title it will use it, if not it will ask the user for the title.
	taskTitle := c.Param("title")
	if taskTitle == "" {
		var taskToDelete struct {
			TitleToDelete string `json:"titleToDelete" binding:"required"`
		}
		if err := c.ShouldBindJSON(&taskToDelete); err != nil {
			log.Printf("Error binding request data: %v", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read title of the task you want to delete",
			})
			return
		}
		taskTitle = taskToDelete.TitleToDelete
	}

	// Attempt to delete the task with the given title and user ID
	result := initializers.DB.Unscoped().Where("title = ? AND user_id = ?", taskTitle, getUserId).Delete(&models.Task{})

	if result.Error != nil {
		log.Printf("Error trying to delete information from the data base: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete task",
		})
		return
	}

	// Check if any rows were affected by the delete operation
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}

func UpdateTask(c *gin.Context) {
	user, _ := c.Get("user") //getting the user information
	userObj := user.(models.User)
	getUserId := userObj.ID
	//Information about which task/ what in the task to update
	var updateFields struct {
		CurrentTitle   string `json:"titleToUpdate" binding:"required"`
		NewTitle       string `json:"newTitle"`
		NewDescription string `json:"newDescription"`
		NewDay         string `json:"newDay"`
		NewMonth       string `json:"newMonth"`
		NewYear        string `json:"newYear"`
	}

	if err := c.ShouldBindJSON(&updateFields); err != nil {
		log.Printf("Error binding request data: %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read JSON data",
		})
		return
	}
	var task models.Task
	//checking if the task exists
	result := initializers.DB.Where("user_id = ? AND title = ?", userObj.ID, updateFields.CurrentTitle).First(&task)
	if result.Error != nil {
		log.Printf("Title not founds: %v", result.Error)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The title you want to update doesn't exist",
		})
		return
	}
	oldDueDate := task.DueDate
	parts := strings.Split(oldDueDate, "/")

	updatedDate := map[string]int{}
	// updating the parts of the date that changed
	if updateFields.NewDay == "" {
		updatedDate["newDay"], _ = strconv.Atoi(parts[1])
	} else {
		updatedDate["newDay"], _ = strconv.Atoi(updateFields.NewDay)
	}
	if updateFields.NewMonth == "" {
		updatedDate["newMonth"], _ = strconv.Atoi(parts[0])
	} else {
		updatedDate["newMonth"], _ = strconv.Atoi(updateFields.NewMonth)
	}
	if updateFields.NewYear == "" {
		updatedDate["newYear"], _ = strconv.Atoi(parts[2])
	} else {
		updatedDate["newYear"], _ = strconv.Atoi(updateFields.NewYear)
	}

	newDueDate := validateDate(updatedDate["newDay"], updatedDate["newMonth"], updatedDate["newYear"], c)
	if newDueDate == "" { // if validateDate returned empty string it means that invalid date was entered
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}
	// Create a map to store the fields to be updated
	fieldsToUpdate := make(map[string]interface{})
	//checking which parts in the task were updated.
	if updateFields.NewTitle != "" {
		fieldsToUpdate["Title"] = updateFields.NewTitle
	}
	if updateFields.NewDescription != "" {
		fieldsToUpdate["Description"] = updateFields.NewDescription
	}
	fmt.Println(newDueDate)
	if newDueDate != "" {
		fieldsToUpdate["DueDate"] = newDueDate
	}
	//updating the data base with the new information
	result = initializers.DB.Model(&models.Task{}).Where("title = ? AND user_id = ?", updateFields.CurrentTitle, getUserId).Updates(fieldsToUpdate)

	if result.Error != nil {
		log.Printf("Error updating the task information %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}
	if result.RowsAffected == 0 {
		log.Printf("Information of the task failed to update")
		c.JSON(http.StatusNotFound, gin.H{"error": "Title not found"})
		return
	}
}

func ReadAllTasks(c *gin.Context) {
	user, _ := c.Get("user")
	userObj := user.(models.User)
	getUserId := userObj.ID
	var tasks []models.Task
	initializers.DB.Where("user_id = ?", getUserId).Find(&tasks)

	c.HTML(http.StatusOK, "readTasks.html", gin.H{
		"AllTasks": tasks,
	})
}

// validate that the Date entered is valid
func validateDate(day, month, year int, c *gin.Context) string {
	dueDate := ""
	if month < 10 {
		dueDate += "0" + strconv.Itoa(month) + "/"
	} else {
		dueDate += strconv.Itoa(month) + "/"
	}
	if day < 10 {
		dueDate += "0" + strconv.Itoa(day) + "/"
	} else {
		dueDate += strconv.Itoa(day) + "/"
	}

	dueDate += strconv.Itoa(year)
	_, err := time.Parse("01/02/2006", dueDate)

	if err != nil {
		log.Printf("Failed to validate the date: %v", err.Error())
		return ""
	}
	return dueDate
}
