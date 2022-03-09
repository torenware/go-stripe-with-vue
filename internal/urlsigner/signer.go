package urlsigner

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
	"golang.org/x/crypto/bcrypt"
)

type Signer struct {
	Secret []byte
}

type HashableInfo struct {
	Payload string `json:"payload"`
	Salt    string `json:"salt"`
}

func (s *Signer) GetHashWithSalt(payload string) (string, error) {
	info := HashableInfo{
		Payload: payload,
		Salt:    string(s.Secret),
	}
	raw, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword(raw, 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *Signer) ConfirmHashForString(hashed, payload string) error {
	info2test := HashableInfo{
		Payload: payload,
		Salt:    string(s.Secret),
	}
	raw, err := json.Marshal(info2test)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashed), raw)
	return err
}

func (s *Signer) GenerateTokenFromString(data string) string {
	var urlToSign string

	crypt := goalone.New(s.Secret, goalone.Timestamp)
	if strings.Contains(data, "?") {
		urlToSign = fmt.Sprintf("%s&hash=", data)
	} else {
		urlToSign = fmt.Sprintf("%s?hash=", data)
	}

	tokenBytes := crypt.Sign([]byte(urlToSign))
	token := string(tokenBytes)
	return token
}

func (s *Signer) VerifyToken(token string) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	_, err := crypt.Unsign([]byte(token))

	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func (s *Signer) Expired(token string, minutesUntilExpire int) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	ts := crypt.Parse([]byte(token))

	return time.Since(ts.Timestamp) > time.Duration(minutesUntilExpire)*time.Minute
}
