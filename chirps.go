package main

import (
  "fmt"
  "strings"
  "net/http"
  "encoding/json"
  "github.com/kekekekyle/database"
  "strconv"
)

func (cfg *apiConfig) handleGetChirpById (w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  chirpId := r.PathValue("chirpId")
  if chirpId == "" {
    w.WriteHeader(404)
    return
  }

  chirps, err := cfg.database.GetChirps()
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
  }

  chirpIndex, err := strconv.Atoi(chirpId)
  if err != nil {
    w.WriteHeader(400)
    return
  }

  var foundChirp database.Chirp
  for _, chirp := range chirps {
    if chirp.Id == chirpIndex {
      foundChirp = chirp
    }
  }

  if foundChirp.Id == 0 {
    w.WriteHeader(404)
    return 
  }

  data, err := json.Marshal(foundChirp)
  if err != nil {
    w.WriteHeader(500)
  }

  w.Write(data)
}

func (cfg *apiConfig) handleGetChirps (w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  chirps, err := cfg.database.GetChirps()
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
  }

  data, err := json.Marshal(chirps)
  if err != nil {
    w.WriteHeader(500)
  }

  w.Write(data)
}

func (cfg *apiConfig) handleCreateChirps (w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  decoder := json.NewDecoder(r.Body)
  chirp := database.Chirp{}
  err := decoder.Decode(&chirp)

  if err != nil || len(chirp.Body) > 140 {
    w.WriteHeader(400)
    w.Write([]byte(`{
      "error": "Something went wrong"
    }`))
    return
  }

  words := strings.Split(chirp.Body, " ")
  cleanedWords := []string{}
  for _, word := range words {
    cleanedWord := word
    if strings.ToLower(word) == "kerfuffle" {
      cleanedWord = "****"
    }
    if strings.ToLower(word) == "sharbert" {
      cleanedWord = "****"
    }
    if strings.ToLower(word) == "fornax" {
      cleanedWord = "****"
    }
    cleanedWords = append(cleanedWords, cleanedWord)
  }
  cleanedString := strings.Join(cleanedWords, " ")

  createdChirp, err := cfg.database.CreateChirp(cleanedString)
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
  }

  data, err := json.Marshal(createdChirp)
  if err != nil {
    w.WriteHeader(500)
  }

  w.WriteHeader(201)
  w.Write(data)
}
