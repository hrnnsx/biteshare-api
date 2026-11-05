package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword converts a plain-text password into a bcrypt hash.
// The plain password should never be stored in the database.
func HashPassword(password string) (string, error) {
	// GenerateFromPassword creates a salted hash using bcrypt's default cost.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// ComparePassword checks whether a plain password matches a stored bcrypt hash.
func ComparePassword(hashedPassword string, password string) error {
	// CompareHashAndPassword returns nil only when the passwords match.
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
