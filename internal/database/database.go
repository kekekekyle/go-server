package database

import (
  "fmt"
  "os"
  "sync"
  "encoding/json"
  "sort"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type User struct {
  Id int `json:"id"`
  Email string `json:"email"`
  Password string `json:"password"`
  ExpiresInSeconds int `json:"expires_in_seconds"`
  Token string `json:"-"`
}

type Chirp struct {
  Id int `json:"id"`
  Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users map[int]User `json:"users"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
  mux := &sync.RWMutex{}

  db := &DB{
    path: path,
    mux: mux,
  }

  err := db.ensureDB()
  if err != nil {
    return nil, err
  }

  return db, nil
}

func (db *DB) FindUserById(id int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

  for _, dbUser := range dbStructure.Users {
    if dbUser.Id == id {
      return dbUser, nil
    }
  }
  return User{}, fmt.Errorf("No user found with id: %v", id)
}


func (db *DB) FindUser(user User) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

  for _, dbUser := range dbStructure.Users {
    if dbUser.Email == user.Email {
      return dbUser, nil
    }
  }
  return User{}, err
}

func (db *DB) UpdateUser(user User) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	dbStructure.Users[user.Id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) CreateUser(user User) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

  existingUser, err := db.FindUser(user)
  if err != nil {
    return User{}, err
  }
  if (User{}) != existingUser {
    return User{}, fmt.Errorf("User already exists")
  }

	id := len(dbStructure.Users) + 1
  user.Id = id
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		Id: id,
		Body: body,
	}
	dbStructure.Chirps[id] = chirp

	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
  dbStructure, err := db.loadDB()
  if err != nil {
    return []Chirp{}, err
  }

  chirps := []Chirp{}
  for _, chirp := range dbStructure.Chirps {
    chirps = append(chirps, chirp)
  }

  sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id})

  return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
  _, err := os.ReadFile(db.path)
  if err != nil {
    os.WriteFile(db.path, []byte(`{"chirps": {}, "users": {}}`), 0666)
  }
  return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
  data, err := os.ReadFile(db.path)
  if err != nil {
    return DBStructure{}, err
  }

  var dbStructure DBStructure
  if err := json.Unmarshal(data, &dbStructure); err != nil {
    return DBStructure{}, err
  }

  return dbStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
  data, err := json.Marshal(dbStructure)
  if err != nil {
    return err
  }

  if err := os.WriteFile(db.path, []byte(data), 0666); err != nil {
    return err
  }
  return nil
}

