package utils

import "golang.org/x/crypto/bcrypt"

// HashCost is the bcrypt cost factor used by HashPassword. Production keeps the
// secure default (12); test suites lower it via a TestMain because cost-12
// hashing takes several seconds per call under the race detector and the
// services/handlers packages create many test users — at cost 12 the race suite
// exceeds the 10-minute per-package test timeout. Lowering it does not affect
// verification: CheckPassword reads the cost embedded in the stored hash.
var HashCost = 12

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), HashCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
