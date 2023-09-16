package dao

import "gorm.io/gorm"

type BaseModel = gorm.Model

type BaseTemplateModel struct {
	BaseModel
	Name string `gorm:"size:256"`
}
