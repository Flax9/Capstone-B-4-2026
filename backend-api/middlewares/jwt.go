package middlewares

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected adalah tameng (interseptor) rute yang menuntut Token JWT Valid
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Menarik dokumen otorisasi dari Header Permintaan HTTP
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Akses Ditolak: Anda mencoba mendobrak jalur kritis tanpa Token Otorisasi (Bearer)",
			})
		}

		// 2. Pemisahan format standar industri "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Format Otorisasi Ilegal: Pastikan header berformat 'Bearer <token>'",
			})
		}
		tokenString := parts[1]

		// 3. Pengolahan Verifikasi Kriptografik (Nol-Latensi Database)
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			secretKey = "capstone_rahasia_negara_2026"
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			// Secara eksplisit melindungi mesin dari serangan Downgrade Algorithm (Misal diubah ke RS256/None)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("serangan algoritma terdeteksi: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Sertifikat JWT kedaluwarsa atau tanda-tangan kriptografi Anda palsu/lecek",
			})
		}

		// 4. Meletakkan KTP/Identitas nasabah asli ke memori sementara Fiber (Bisa diakses oleh rute di bawahnya)
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Locals("user_id", claims["sub"])
			c.Locals("username", claims["username"])
		}

		// TANDA TANGAN ASLI: Silakan Fiber melaju ke fungsi utama (Get Balance / Transfer)
		return c.Next()
	}
}
