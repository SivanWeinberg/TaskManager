package models

import "gorm.io/gorm"

type Task struct {
	gorm.Model
	Title       string `gorm:"uniqueIndex:unique_task_title_user_id"`
	Description string
	DueDate     string
	UserId      uint `gorm:"uniqueIndex:unique_task_title_user_id"`
}
