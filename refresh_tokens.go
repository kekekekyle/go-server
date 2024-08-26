package main

import (
  "fmt"
  "time"
  "encoding/json"
  "net/http"
  "strings"
  "crypto/rand"
  "encoding/hex"
  "github.com/kekekekyle/database"
)

func (cfg *apiConfig) createRefreshToken () (database.RefreshToken, error) {
  var refreshToken database.RefreshToken
	c := 32
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return refreshToken, err
	}

  refreshToken.RefreshToken = hex.EncodeToString(b)
  refreshToken.ExpiresAt = int(time.Now().AddDate(0, 0, 60).Unix())

  return refreshToken, nil
}

func (cfg *apiConfig) handleRevokeToken (w http.ResponseWriter, r *http.Request) {
  authorizationHeader := r.Header.Get("Authorization")
  refreshToken := strings.Split(authorizationHeader, "Bearer ")[1]

  if err := cfg.database.DeleteRefreshToken(refreshToken); err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return 
  }
  w.WriteHeader(204)
}

func (cfg *apiConfig) handleRefreshToken (w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  authorizationHeader := r.Header.Get("Authorization")
  refreshToken := strings.Split(authorizationHeader, "Bearer ")[1]

  user, err := cfg.database.FindRefreshToken(refreshToken)
  if err != nil {
    w.WriteHeader(401)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  if int(time.Now().Unix()) > user.RefreshToken.ExpiresAt {
    w.WriteHeader(401)
    w.Write([]byte("Refresh token is expired"))
    return
  }

  jwtToken, err := cfg.createJWT(user)

  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte("Something went wrong"))
  }

  type returnVals struct {
    Token string `json:"token"`
  }

  data, err := json.Marshal(returnVals{Token: jwtToken})
  if err != nil {
    w.WriteHeader(500)
    return
  }

  w.WriteHeader(200)
  w.Write(data)
}

