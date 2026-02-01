/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controlplane

import (
	"crypto/rsa"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token generation and verification
type JWTManager struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	logger     *slog.Logger
}

// NewJWTManager creates a new JWT manager and loads keys from files
func NewJWTManager(logger *slog.Logger) (*JWTManager, error) {
	jm := &JWTManager{
		logger: logger,
	}

	// Load public key
	publicKey, err := loadPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load JWT public key: %w", err)
	}
	jm.publicKey = publicKey

	// Load private key
	privateKey, err := loadPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load JWT private key: %w", err)
	}
	jm.privateKey = privateKey

	logger.Info("JWT manager initialized successfully")
	return jm, nil
}

// loadPublicKey loads the RSA public key from file
func loadPublicKey() (*rsa.PublicKey, error) {
	pubKeyData, err := os.ReadFile(jwtPublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return pubKey, nil
}

// loadPrivateKey loads the RSA private key from file
func loadPrivateKey() (*rsa.PrivateKey, error) {
	privKeyData, err := os.ReadFile(jwtPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privKey, nil
}

// TokenClaims represents the JWT claims
type TokenClaims struct {
	jwt.RegisteredClaims
	Subject string
}

// GenerateToken generates a new JWT token with no expiry
func (jm *JWTManager) GenerateToken(subject string, customClaims map[string]interface{}) (string, error) {
	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: subject,
		},
		Subject: subject,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign the token with the private key
	tokenString, err := token.SignedString(jm.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	jm.logger.Info("JWT token generated", "subject", subject)
	return tokenString, nil
}

// VerifyToken verifies a JWT token and returns the claims
func (jm *JWTManager) VerifyToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
