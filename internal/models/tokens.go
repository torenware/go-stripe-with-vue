package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"time"
)

const (
	ScopeAuthentication = "authentication"
)

type Token struct {
	ID        int       `json:"id"`
	PlainText string    `json:"token"`
	UserID    int64     `json:"-"`
	Email     string    `json:"-"`
	Name      string    `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func GenerateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	token := Token{
		UserID: int64(userID),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]
	return &token, nil
}

func (m *DBModel) InsertToken(token *Token, user User) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Delete any stray tokens still hanging out.
	delete := `
	delete from tokens where user_id = ?
	`
	_, err := m.DB.ExecContext(ctx, delete, user.ID)
	if err != nil {
		return err
	}

	stmt := `
		insert into tokens
			(user_id, name, email, token_hash,
			 created_at, updated_at)
		values (?, ?, ?, ?, ?, ?)
	`

	_, err = m.DB.ExecContext(ctx, stmt,
		user.ID,
		user.LastName,
		user.Email,
		token.Hash,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}
	return nil
}

func CreateTokenHash(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func (m *DBModel) GetUserFromToken(token string, ttl time.Duration) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	hash := CreateTokenHash(token)
	var created time.Time
	var u User

	query := `
		select
			u.id, u.first_name, u.last_name,
			u.email, t.created_at
		from users u
		inner join tokens t on t.user_id = u.id
		where t.token_hash = ?
	`
	row := m.DB.QueryRowContext(ctx, query, hash)
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&created,
	)

	if err != nil {
		return nil, err
	}

	expired := created.Add(ttl)
	if expired.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	return &u, nil
}

func (m *DBModel) GetEmailFromToken(token string, ttl time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	hash := CreateTokenHash(token)
	var created time.Time
	var t Token

	row := m.DB.QueryRowContext(ctx, `
		select
			email, created_at
		from tokens
		where token_hash = ?
	`, hash)
	err := row.Scan(
		&t.Email,
		&created,
	)
	if err != nil {
		return "", err
	}
	expires := created.Add(ttl)
	if time.Now().Before(expires) {
		return t.Email, nil
	}
	return "", errors.New("token expired")
}
