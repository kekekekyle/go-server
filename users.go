package main

import (
  "fmt"
  "net/http"
  "encoding/json"
  "strings"
  "strconv"
  "github.com/kekekekyle/database"
  "golang.org/x/crypto/bcrypt"
  "github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handleUpdateUser (w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  authorizationHeader := r.Header.Get("Authorization")
  token := strings.Split(authorizationHeader, "Bearer ")[1]

  keyFunc := func(token *jwt.Token) (interface{}, error){
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
      return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
    }
    return []byte(cfg.jwtSecret), nil
  }

  type userClaims struct {
    jwt.RegisteredClaims
  }

  claims := &userClaims{}

  parsedToken, err := jwt.ParseWithClaims(
    token,
    claims, 
    keyFunc,
  )
  if err != nil {
    w.WriteHeader(401)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  decoder := json.NewDecoder(r.Body)
  user := database.User{}
  if err = decoder.Decode(&user); err != nil {
    w.WriteHeader(400)
    w.Write([]byte(`{
      "error": "Something went wrong"
    }`))
    return
  }

  if claims, ok := parsedToken.Claims.(*userClaims); ok && parsedToken.Valid {
    userId, err := strconv.Atoi(claims.Subject)
    if err != nil {
      w.WriteHeader(401)
      w.Write([]byte(fmt.Sprintf("%v", err)))
    }

    foundUser, err := cfg.database.FindUserById(userId)
    if err != nil {
      w.WriteHeader(401)
      w.Write([]byte(fmt.Sprintf("%v", err)))
    }

    foundUser.Email = user.Email
    hashedPassword, err := bcrypt.GenerateFromPassword(
      []byte(user.Password),
      bcrypt.DefaultCost,
    )
    foundUser.Password = string(hashedPassword)
    updateUser, err := cfg.database.UpdateUser(foundUser)
    if err != nil {
      w.WriteHeader(500)
      w.Write([]byte(fmt.Sprintf("%v", err)))
      return
    }

    type returnedUser struct {
      Id int `json:"id"`
      Email string `json:"email"`
      RefreshToken string `json:"refresh_token"`
    }
    returnUser := returnedUser{
      Id: updateUser.Id,
      Email: updateUser.Email,
      RefreshToken: updateUser.RefreshToken.RefreshToken,
    }

    data, err := json.Marshal(returnUser)
    if err != nil {
      w.WriteHeader(500)
      return
    }

    w.WriteHeader(200)
    w.Write(data)
    return
  } else {
    w.WriteHeader(401)
    w.Write([]byte(`{
      "error": "Something went wrong"
    }`))
    return
  }
}


func (cfg *apiConfig) handleCreateUser (w http.ResponseWriter, r *http.Request) {
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

  hashedPassword, err := bcrypt.GenerateFromPassword(
    []byte(user.Password),
    bcrypt.DefaultCost,
  )
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
  }
  user.Password = string(hashedPassword)

  createdUser, err := cfg.database.CreateUser(user)
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
  }

  type returnedUser struct {
    Id int `json:"id"`
    Email string `json:"email"`
    Password string `json:"-"`
    ExpiresInSeconds int `json:"-"`
    Token string `json:"-"`
    RefreshToken database.RefreshToken `json:"-"`
    IsChirpyRed bool `json:"is_chirpy_red"`
  }

  data, err := json.Marshal(returnedUser(createdUser))
  if err != nil {
    w.WriteHeader(500)
  }

  w.WriteHeader(201)
  w.Write(data)
}

