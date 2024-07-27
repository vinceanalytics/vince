package models

import (
	"fmt"

	"github.com/gernest/len64/web/db/schema"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Create a user and API key. No validation is done. If A User exists no further
// operations are done.
func Bootstrap(
	db *gorm.DB,
	name, email, password, key string,
) error {
	if schema.Exists(db, func(db *gorm.DB) *gorm.DB {
		return db.Model(&schema.User{}).Where("email = ?", email)
	}) {
		return nil
	}
	hashPasswd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password%w", err)
	}
	u := &schema.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashPasswd),
	}
	err = db.Create(u).Error
	if err != nil {
		return fmt.Errorf("saving bootstrapped user%w", err)
	}
	return nil
}
