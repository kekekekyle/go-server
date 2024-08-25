package main

import (
  "fmt"
  "time"
  "net/http"
  "encoding/json"
  "strconv"
  "github.com/kekekekyle/database"
  "golang.org/x/crypto/bcrypt"
  "github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handleLogin (w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  decoder := json.NewDecoder(r.Body)
  user := database.User{}
  err := decoder.Decode(&user)
  if err != nil {
    w.WriteHeader(400)
    w.Write([]byte(`{
      "error": "Something went wrong"
    }`))
    return
  }

  foundUser, err := cfg.database.FindUser(user)
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
  }
  if (database.User{}) == foundUser {
    w.WriteHeader(400)
    w.Write([]byte(`{
      "error": "User does not exist"
    }`))
  }

  err = bcrypt.CompareHashAndPassword(
    []byte(foundUser.Password),
    []byte(user.Password),
  )
  if err != nil {
    w.WriteHeader(401)
    w.Write([]byte(fmt.Sprintf("%v", err)))
  }

  if user.ExpiresInSeconds == 0 || user.ExpiresInSeconds > 24 * 60 * 60 {
    user.ExpiresInSeconds = 24 * 60 * 60
  }

  currentTime := time.Now()
  issuedAt := jwt.NumericDate{
    Time: currentTime,
  }
  expiresAt := jwt.NumericDate{
    Time: currentTime.Add(time.Second * time.Duration(user.ExpiresInSeconds)),
  }

  jwtClaims := jwt.RegisteredClaims{
    Issuer: "chirpy",
    IssuedAt: &issuedAt,
    ExpiresAt: &expiresAt,
    Subject: strconv.Itoa(foundUser.Id),
  }
  jwt := jwt.NewWithClaims(
    jwt.SigningMethodHS256,
    jwtClaims,
  )

  signedString, err := jwt.SignedString([]byte(cfg.jwtSecret))
  if err != nil {
    w.WriteHeader(500)
  }

  foundUser.Token = signedString

  type returnedUser struct {
    Id int `json:"id"`
    Email string `json:"email"`
    Password string `json:"-"`
    ExpiresInSeconds int `json:"-"`
    Token string `json:"token"`
  }

  data, err := json.Marshal(returnedUser(foundUser))
  if err != nil {
    w.WriteHeader(500)
  }

  w.Write(data)
}

