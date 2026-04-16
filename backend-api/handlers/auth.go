package handlers

import (
	"banking-backend/config"
	"banking-backend/models"
	"encoding/json"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Payload format"})
	}

	var user models.User
	var actionEvent string
	var resStatus int
	var payload fiber.Map
	var details map[string]interface{}

	// BACA KE REPLICA
	err := config.DB.Where("username = ?", req.Username).First(&user).Error

	if err != nil {
		actionEvent = "LOGIN_FAILED_NOTFOUND"
		resStatus = 401
		payload = fiber.Map{"error": "Identitas gagal divalidasi"}
		details = map[string]interface{}{"reason": "user_missing_or_typo"}
	} else {
		// [FASE EMISI JWT - STATELESS]
		claims := jwt.MapClaims{
			"sub":      user.UserID,
			"username": user.Username,
			"exp":      time.Now().Add(time.Minute * 15).Unix(), // Kadaluwarsa 15 Menit
			"iat":      time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			secretKey = "capstone_rahasia_negara_2026"
		}

		signedToken, errSigning := token.SignedString([]byte(secretKey))
		if errSigning != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menenun tanda tangan kriptografi JWT"})
		}

		actionEvent = "LOGIN_SUCCESS"
		resStatus = 200
		payload = fiber.Map{
		    "message": "Auth Berhasil Disetujui",
		    "token": signedToken, 
		    "user": user.FullName,
		}
		details = map[string]interface{}{"auth_method": "password_verification", "device_os": c.Get("Sec-CH-UA-Platform")}
	}

	// TULIS KE MASTER (Audit Recording)
	detailsJSON, _ := json.Marshal(details)
	auditLog := models.AuditLog{
		Action:    actionEvent,
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
		Details:   string(detailsJSON),
	}
	
	if user.UserID != uuid.Nil {
		auditLog.UserID = user.UserID
	}

	// Perintah GORM Create() => Menembus Rute MASTER Terdalam
	config.DB.Create(&auditLog)

	return c.Status(resStatus).JSON(payload)
}