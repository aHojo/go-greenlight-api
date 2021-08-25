package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"crypto/rand"
	"time"

	"github.com/ahojo/greenlight/internal/validator"
)

// Define constants for the token scope
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

// Token hold the data for a token
// Includes the plaintext and hashed versions of the token
type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(UserID int64, ttl time.Duration, scope string) (*Token, error) {

	token := &Token{
		UserID: UserID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Initialize a zero-valued byte slice with a length of 16 bytes
	randomBytes := make([]byte, 16)

	// Read() function to fill the byte slice with random bytes from the OS's CSPRNG
	// Err if CSPRNG fails to function correctly
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to a base-32-encoded string and assign it to the token Plaintext
	// This string will be sent in the user email
	// Looks like
	//
	// QMGX3PJ3WLRL2YRTQGQ6KRHU
	// By Default base-32 strings may be padded at the end with =
	// We do not need this so we use WithPadding(base32.NoPadding) to omit them
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Generate the SHA-256 hash of the plaintext token string
	// We store this in the database
	// Note that the
	// sha256.Sum256() function returns an *array* of length 32, so to make it easier to
	// work with we convert it to a slice using the [:] operator before storing it
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil

	/*
			The length of the plaintext token string itself depends on how those 16 random bytes are
		  encoded to create a string. In our case we encode the random bytes to a base-32 string,
		  which results in a string with 26 characters. In contrast, if we encoded the random bytes
		  using hexadecimal (base-16) the string would be 32 characters long instead.
	*/
}

/* TOKEN MODEL and VALIDATION */

// Check that the plaintext token has been provided and is exactly 26 bytes long
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

// New() shortcut to create a new Token struct and insert the data into the database
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {

	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

// Insert adds the data for a specific token to the tokens table
func (m TokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4);
	`
	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// DeleteAllForUser deletes all tokens for a specific user and scope
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {

	query := `
		DELETE FROM tokens
		WHERE scope = $1 AND user_id = $2;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
