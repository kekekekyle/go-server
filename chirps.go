package main

import (
  "fmt"
  "strings"
  "net/http"
  "encoding/json"
  "github.com/kekekekyle/database"
  "strconv"
  "sort"
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
  qAuthorId := r.URL.Query().Get("author_id")

  qSort := r.URL.Query().Get("sort")

  chirps, err := cfg.database.GetChirps()
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
  }

  if qSort == "asc" {
    sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id})
  }

  if qSort == "desc" {
    sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id > chirps[j].Id})
  }

  if qAuthorId != "" {
    authorId, err := strconv.Atoi(qAuthorId)
    if err != nil {
      w.WriteHeader(500)
      w.Write([]byte(fmt.Sprintf("%v", err)))
    }
    filteredChirps := []database.Chirp{}
    for _, chirp := range chirps {
      if chirp.AuthorId == authorId {
        filteredChirps = append(filteredChirps, chirp)
      }
    }
    data, err := json.Marshal(filteredChirps)
    if err != nil {
      w.WriteHeader(500)
    }
    w.Write(data)
  }
  data, err := json.Marshal(chirps)
  if err != nil {
    w.WriteHeader(500)
  }
  w.Write(data)
}

type HandleCreateChirps struct {
  api *apiConfig
}

func (h *HandleCreateChirps) ServeHTTP (w http.ResponseWriter, r *http.Request) {
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

  userHeader := r.Header.Get("User")
  user := database.User{}
  if err := json.Unmarshal([]byte(userHeader), &user); err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  createdChirp, err := h.api.database.CreateChirp(cleanedString, user.Id)
  if err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  data, err := json.Marshal(createdChirp)
  if err != nil {
    w.WriteHeader(500)
    return
  }

  w.WriteHeader(201)
  w.Write(data)
}

type HandleDeleteChirps struct {
  api *apiConfig
}

func (h *HandleDeleteChirps) ServeHTTP (w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  chirpId := r.PathValue("chirpId")
  if chirpId == "" {
    w.WriteHeader(404)
    return
  }

  chirps, err := h.api.database.GetChirps()
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

  userHeader := r.Header.Get("User")
  user := database.User{}
  if err := json.Unmarshal([]byte(userHeader), &user); err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  if foundChirp.AuthorId != user.Id {
    w.WriteHeader(403)
  }

  if err = h.api.database.DeleteChirp(chirpIndex); err != nil {
    w.WriteHeader(500)
    w.Write([]byte(fmt.Sprintf("%v", err)))
    return
  }

  w.WriteHeader(204)
}
