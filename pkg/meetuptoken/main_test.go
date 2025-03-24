package meetuptoken

import (
	"context"
	"path/filepath"
	"testing"
)

func TestGetToken(t *testing.T) {
	//dir, _ := os.Getwd()
	config := LoadConfigFromEnvFile(filepath.Join("..", ".."), ".env")

	token, err := GetToken(context.Background(), config, "test")

	if err != nil {
		t.Fatalf("Expcted no error, got: %v", err)
	}

	if token == "" {
		t.Error("No token retrieved")
	}
}
