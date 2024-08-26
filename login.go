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

func (cfg *apiConfig) createJWT (user database.User) (string, error) {
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
    Subject: strconv.Itoa(user.Id),
  }
  jwt := jwt.NewWithClaims(
    jwt.SigningMethodHS256,
    jwtClaims,
  )

  signedString, err := jwt.SignedString([]byte(cfg.jwtSecret))
  if err != nil {
    return "", err
  }
  return signedString, nil
}

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
    return
  }
  if (database.User{}) == foundUser {
    w.WriteHeader(400)
    w.Write([]byte(`{
      "error": "User does not exist"
    }`))
    return
  }

  err = bcrypt.CompareHashAndPassword(
    []byte(foundUser.Password),
    []byte(user.Password),
  )
  if err != nil {
    w.WriteHeader(401)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  if user.ExpiresInSeconds == 0 || user.ExpiresInSeconds > 3600 {
    user.ExpiresInSeconds = 3600
  }

  user.Id = foundUser.Id
  foundUser.ExpiresInSeconds = user.ExpiresInSeconds

  signedString, err := cfg.createJWT(user)
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }
  foundUser.Token = signedString

  refreshToken, err := cfg.createRefreshToken()
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("panic here? %v", err)))
    return
  }

  createdRefreshToken, err := cfg.database.CreateRefreshToken(user, refreshToken)
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  foundUser.RefreshToken = createdRefreshToken

  updateUser, err := cfg.database.UpdateUser(foundUser)
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  type returnedUser struct {
    Id int `json:"id"`
    Email string `json:"email"`
    Token string `json:"token"`
    RefreshToken string `json:"refresh_token"`
    IsChirpyRed bool `json:"is_chirpy_red"`
  }
  returnUser := returnedUser{
    Id: updateUser.Id,
    Email: updateUser.Email,
    Token: updateUser.Token,
    RefreshToken: updateUser.RefreshToken.RefreshToken,
    IsChirpyRed: updateUser.IsChirpyRed,
  }

  data, err := json.Marshal(returnUser)
  if err != nil {
    w.WriteHeader(500)
    return
  }

  w.Write(data)
}

