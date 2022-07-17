package template

import "text/template"

var ModelMap = map[string]*template.Template{
	"model": template.Must(template.New("model").Parse(Models())),
}

// Models projectDir/model/model.go
func Models() string {
	return `
package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	return nil
}
`
}
