package main

import (
  "fmt"
  "encoding/json"
  "net/http"
  "strconv"
  "strings"
  "github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) authenticate (next http.Handler) http.Handler {
  nextHandler := func (w http.ResponseWriter, r *http.Request) {
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

    if claims, ok := parsedToken.Claims.(*userClaims); ok && parsedToken.Valid {
      userId, err := strconv.Atoi(claims.Subject)
      if err != nil {
        w.WriteHeader(401)
        w.Write([]byte(fmt.Sprintf("%v", err)))
        return
      }

      foundUser, err := cfg.database.FindUserById(userId)
      if err != nil {
        w.WriteHeader(401)
        w.Write([]byte(fmt.Sprintf("%v", err)))
        return
      }

      headerUser, err := json.Marshal(foundUser)
      if err != nil {
        w.WriteHeader(500)
        w.Write([]byte(fmt.Sprintf("%v", err)))
        return
      }

      r.Header.Set("User", string(headerUser))
      next.ServeHTTP(w, r)
    } else {
      w.WriteHeader(500)
      w.Write([]byte("Something went wrong"))
    }
  }
  return http.HandlerFunc(nextHandler)
}

