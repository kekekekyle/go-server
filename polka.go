package main

import (
  "fmt"
  "net/http"
  "encoding/json"
  "strings"
)

type PolkaData struct {
  UserId int `json:"user_id"`
}

type Polka struct {
  Event string `json:"event"`
  Data PolkaData `json:"data"`
}

func (cfg *apiConfig) handlePolkaWebhooks (w http.ResponseWriter, r *http.Request) {
  authorizationHeader := r.Header.Get("Authorization")
  if authorizationHeader == "" {
    w.WriteHeader(401)
    return
  }
  apiKey := strings.Split(authorizationHeader, "ApiKey ")[1]
  if apiKey != cfg.polkaApiKey {
    w.WriteHeader(401)
    return
  }

  decoder := json.NewDecoder(r.Body)
  polka := Polka{}
  if err := decoder.Decode(&polka); err != nil {
    w.WriteHeader(400)
    w.Write([]byte(`{
      "error": "Something went wrong"
    }`))
    return
  }

  if polka.Event != "user.upgraded" {
    w.WriteHeader(204)
    return
  }

  user, err := cfg.database.FindUserById(polka.Data.UserId)
  if err != nil {
    w.WriteHeader(404)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  user.IsChirpyRed = true
  _, err = cfg.database.UpdateUser(user)
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  w.WriteHeader(204)
}
