package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type IUserModel interface {
	Insert(user *User) error
	GetByEmail(email string) (*User, error)
	UpdatePassword(userID int, password string) error
	DeleteUser(userEmail string) error
}

type User struct {
	ID        int64     `json:"id"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserModel struct {
	DB *sql.DB
}

type UserModelMock struct {
	DB []*User
}

func EncryptPassword(plaintext string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", err

	}
	return string(hashedBytes), nil
}

func (m *UserModel) Insert(user *User) error {
	hashedPassword, err := EncryptPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	query := `
	INSERT INTO users (email, password_hash, created_at)
	VALUES ($1, $2, $3)
	RETURNING id`

	args := []interface{}{user.Email, user.Password, user.CreatedAt}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "users_email_key"):
			return errors.New("duplicate email")
		default:
			return err
		}
	}

	return nil
}

func (m *UserModel) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id, password_hash, email, created_at
	FROM users
	WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("record not found")
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m *UserModel) UpdatePassword(userID int, password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	passwordHash := string(hashedBytes)
	query := `UPDATE users
	SET password_hash = $2
	WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// we actually only need to check for error, we're going to see the rows affected and return the no data then
	result, err := m.DB.ExecContext(ctx, query, userID, passwordHash)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("no data")
	}

	return nil

}

func (m *UserModel) DeleteUser(userEmail string) error {
	query := `delete from users where email = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// same as above, we need to check result
	result, err := m.DB.ExecContext(ctx, query, userEmail)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("no data")
	}
	return nil
}

func (mockUM *UserModelMock) Insert(user *User) error {
	hashedPassword, err := EncryptPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	targetUser := user
	for _, userToCheck := range mockUM.DB {
		if userToCheck.Email == targetUser.Email {
			return errors.New("duplicate email")
		}
	}
	mockUM.DB = append(mockUM.DB, user)
	return nil
}

func (mockUM *UserModelMock) GetByEmail(email string) (*User, error) {
	for _, user := range mockUM.DB {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("record not found")
}

func (mockUM *UserModelMock) UpdatePassword(userID int, password string) error {
	hashedPassword, err := EncryptPassword(password)
	if err != nil {
		return err
	}
	newpassword := hashedPassword
	for _, user := range mockUM.DB {
		if user.ID == int64(userID) {
			user.Password = newpassword
			return nil
		}
	}
	return errors.New("no data")
}

func (mockUM *UserModelMock) DeleteUser(userEmail string) error {
	for i, user := range mockUM.DB {
		if user.Email == userEmail {
			mockUM.DB = append(mockUM.DB[:i], mockUM.DB[i+1:]...)
			return nil
		}
	}
	return errors.New("no data")
}
