package services

import (
	"os"
	"testing"

	"maltiden/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

// TestMain lowers the bcrypt cost for the whole services test binary. Cost-12
// hashing takes several seconds per call under the race detector, and this
// package creates many test users (createTestUser → AuthService.Register); at
// the default cost the race suite exceeds the 10-minute per-package timeout.
// Production hashing is unaffected — utils.HashCost defaults to 12.
func TestMain(m *testing.M) {
	utils.HashCost = bcrypt.MinCost
	os.Exit(m.Run())
}
