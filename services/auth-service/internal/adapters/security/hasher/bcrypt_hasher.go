package hasher

import (
	"golang.org/x/crypto/bcrypt"
	"bangungrid/services/auth-service/internal/core/ports"
)

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) ports.PasswordHasherPort {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	return string(bytes), err
}

func (h *BcryptHasher) Compare(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
