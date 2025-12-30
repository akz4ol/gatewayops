// Command generate-key generates API keys for GatewayOps.
//
// Usage:
//
//	go run scripts/generate-key.go [env]
//
// Where env is one of: dev, stg, prd (defaults to dev)
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	env := "dev"
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	// Validate environment
	switch env {
	case "dev", "stg", "prd":
		// valid
	default:
		fmt.Fprintf(os.Stderr, "Invalid environment: %s. Must be dev, stg, or prd\n", env)
		os.Exit(1)
	}

	// Generate API key
	apiKey, err := generateAPIKey(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate API key: %v\n", err)
		os.Exit(1)
	}

	// Generate hash for storage
	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), 12)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to hash API key: %v\n", err)
		os.Exit(1)
	}

	// Extract prefix for database lookup
	prefix := apiKey[:12] + apiKey[12:20]

	fmt.Println("=== GatewayOps API Key Generator ===")
	fmt.Println()
	fmt.Printf("Environment: %s\n", env)
	fmt.Println()
	fmt.Printf("API Key (save this, it won't be shown again):\n")
	fmt.Printf("  %s\n", apiKey)
	fmt.Println()
	fmt.Printf("Key Prefix (for database lookup):\n")
	fmt.Printf("  %s\n", prefix)
	fmt.Println()
	fmt.Printf("Bcrypt Hash (store this in database):\n")
	fmt.Printf("  %s\n", string(hash))
	fmt.Println()
	fmt.Println("SQL to insert this key:")
	fmt.Println()
	fmt.Printf(`INSERT INTO api_keys (
    id, org_id, team_id, name, key_prefix, key_hash, environment, permissions, rate_limit_rpm
) VALUES (
    gen_random_uuid(),
    '00000000-0000-0000-0000-000000000001',  -- Default org
    '00000000-0000-0000-0000-000000000001',  -- Default team
    'Development Key',
    '%s',
    '%s',
    '%s',
    'mcp:read,mcp:write',
    1000
);
`, prefix, string(hash), env)
}

func generateAPIKey(env string) (string, error) {
	// Generate 16 random bytes (32 hex chars)
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	randomStr := hex.EncodeToString(randomBytes)
	return "gwo_" + env + "_" + randomStr, nil
}
