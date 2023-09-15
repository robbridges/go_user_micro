package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"the_lonely_road/validator"
	"time"
)

type IUserModel interface {
	Insert(user *User) error
	GetByEmail(email string) (*User, error)
	UpdatePassword(userID int, password string) error
	DeleteUser(userEmail string) error
	Authenticate(email, password string) (*User, error)
	EnterPasswordHash(email, passwordHash, salt string) error
	ConsumePasswordReset(email string) error
}

type User struct {
	ID                     int64     `json:"id"`
	Password               string    `json:"password"`
	Email                  string    `json:"email"`
	CreatedAt              time.Time `json:"created_at"`
	PasswordResetHashToken string    `json:"-"`
	PasswordResetExpiry    time.Time `json:"-"`
	PasswordResetSalt      string    `json:"-"`
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
	INSERT INTO users (email, password_hash, created_at, password_reset_expires, password_reset_token, password_reset_salt)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`

	args := []interface{}{user.Email, user.Password, user.CreatedAt, user.PasswordResetExpiry, user.PasswordResetHashToken, user.PasswordResetSalt}
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
	SELECT id, password_hash, email, created_at, password_reset_expires, password_reset_token, password_reset_salt
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
		&user.PasswordResetExpiry,
		&user.PasswordResetHashToken,
		&user.PasswordResetSalt,
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

func (m *UserModel) EnterPasswordHash(email, passwordHash, salt string) error {
	expiry := time.Now().Add(30 * time.Minute).UTC()
	query := `UPDATE users
	SET password_reset_expires = $1,
		password_reset_token = $2,
		password_reset_salt = $3
	WHERE email = $4
	RETURNING password_reset_expires, password_reset_token, password_reset_salt;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := m.DB.ExecContext(ctx, query, expiry, passwordHash, salt, email)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// ConsumePasswordReset after resetting user password we should consume the token for security reasons
func (m *UserModel) ConsumePasswordReset(email string) error {
	query := `UPDATE users
	SET password_reset_expires = $1,
		password_reset_token = $2,
		password_reset_salt = $3
	WHERE email = $4
	RETURNING password_reset_expires, password_reset_token, password_reset_salt;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := m.DB.ExecContext(ctx, query, time.Time{}, "", "", email)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (*User, error) {
	email = strings.ToLower(email)
	user := User{
		Email: email,
	}

	row := m.DB.QueryRow(
		`SELECT id, password_hash
		FROM users WHERE email=$1`, email,
	)

	err := row.Scan(&user.ID, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return nil, fmt.Errorf("compare() error: %v", err)
	}

	return &user, nil
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

func (mockUM *UserModelMock) Authenticate(email, password string) (*User, error) {
	email = strings.ToLower(email)
	for _, user := range mockUM.DB {
		if user.Email == email {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				return nil, fmt.Errorf("compare() error: %v", err)
			}
			return user, nil
		}
	}
	return nil, errors.New("no data")
}

func (mockUM *UserModelMock) EnterPasswordHash(email, passwordHash, salt string) error {
	user, err := mockUM.GetByEmail(email)
	if err != nil {
		return err
	}
	user.PasswordResetExpiry = time.Now().Add(30 * time.Minute)
	user.PasswordResetHashToken = passwordHash
	user.PasswordResetSalt = salt
	return nil
}

func (mockUM *UserModelMock) ConsumePasswordReset(email string) error {
	user, err := mockUM.GetByEmail(email)
	if err != nil {
		return err
	}
	user.PasswordResetExpiry = time.Time{}
	user.PasswordResetHashToken = ""
	user.PasswordResetSalt = ""
	return nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", email, "must be provided")
	v.Check(len(email) >= 5, "Email", "must be at least 5 bytes long")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "Password", "must be provided")
	v.Check(len(password) >= 4, "Password", "must at least be 4 characters long")
	v.Check(len(password) <= 72, "Password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Email != "", "Email", "must be provided")
	v.Check(len(user.Email) >= 3, "Email", "must be at least 3 bytes long")
	v.Check(len(user.Email) <= 500, "Email", "must be less than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password == "" {
		ValidatePasswordPlaintext(v, user.Password)
	}

}
