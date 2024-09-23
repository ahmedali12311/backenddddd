package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
)

type Envelope map[string]interface{}

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrMissingToken  = errors.New("missing authorization token")
	ErrInvalidClaims = errors.New("invalid token claims")
)

func SendJSONResponse(w http.ResponseWriter, status int, data Envelope) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return err
	}
	return nil
}

var ImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

// IsImageFile checks if the provided filename has an image file extension
func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, validExt := range ImageExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// SaveImageFile saves the uploaded image file to a specified directory with a new name
func SaveImageFile(file io.Reader, table string, filename string) (string, error) {
	if !IsImageFile(filename) {
		return "", fmt.Errorf("file is not a valid image")
	}
	// Create directory structure if it doesn't exist
	fullPath := filepath.Join("uploads", table)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return "", err
	}

	// Generate new filename
	randomNumber := rand.Intn(1000)
	timestamp := time.Now().Unix()
	ext := filepath.Ext(filename)
	newFileName := fmt.Sprintf("%s_%d_%d%s", filepath.Base(table), timestamp, randomNumber, ext)

	newFilePath := filepath.Join(fullPath, newFileName)

	// Save the file
	destFile, err := os.Create(newFilePath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		return "", err
	}

	return newFilePath, nil
}

func HashPassword(password string) (string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashPassword), nil
}

func DeleteImageFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {

		return fmt.Errorf("failed to delete file %s: %w", filePath, err)
	}
	return nil
}

// for converting string to float
func NormalizeFloatInput(input string) string {
	if strings.Contains(input, ".") {
		parts := strings.Split(input, ".")
		if len(parts[1]) == 0 {
			return input + "0"
		}
	}
	return input + ".0"
}

var jwtSecret = []byte("ahmedpa55word")

func GenerateToken(userID, userRole string) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 24).Unix() // 24 hours expiration time

	claims := &jwt.MapClaims{
		"id":       userID,
		"userRole": userRole,
		"exp":      expirationTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	segments := strings.Split(tokenString, ".")
	if len(segments) != 3 {
		return nil, fmt.Errorf("token contains an invalid number of segments")
	}

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
}

func CheckPassword(storedHash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	return err == nil
}

// ParseBoolOrDefault parses a string into a boolean, or returns a default value if parsing fails.
func ParseBoolOrDefault(value string, defaultValue bool) (bool, error) {
	if value == "" {
		return defaultValue, nil
	}
	return strconv.ParseBool(value)
}
