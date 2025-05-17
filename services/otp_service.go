package services

import (
	"github.com/pquerna/otp/totp"
)

// GenerateUserSecret generates a new TOTP secret and provisioning URI for QR code
func GenerateUserSecret(username string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "MyAttendanceApp",
		AccountName: username,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// VerifyTOTP validates a TOTP token using the shared secret
func VerifyTOTP(secret, token string) bool {
	return totp.Validate(token, secret)
}
