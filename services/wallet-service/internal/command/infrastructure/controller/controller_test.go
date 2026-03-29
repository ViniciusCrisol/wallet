package controller

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	godotenv.Load("../../../../.env.test")

	os.Exit(m.Run())
}
