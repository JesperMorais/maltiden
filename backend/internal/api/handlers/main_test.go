package handlers

import (
	"os"
	"testing"

	"maltiden/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

// TestMain lowers the bcrypt cost for the whole handlers test binary. These
// tests register/log in users through the real service stack (cost-12 bcrypt),
// which is several seconds per call under the race detector; lowering the cost
// keeps the package well under the 10-minute test timeout. Production hashing is
// unaffected — utils.HashCost defaults to 12.
func TestMain(m *testing.M) {
	utils.HashCost = bcrypt.MinCost
	os.Exit(m.Run())
}
