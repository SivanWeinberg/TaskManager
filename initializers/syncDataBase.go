package initializers

import "auth/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{}, &models.Task{}) // Creating new tables if they don't exist. If they do exist it will update them if needed.
}
