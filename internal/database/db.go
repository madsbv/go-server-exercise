package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"slices"
	"sync"
)

type Chirp struct {
	Body string `json:"body"`
	Id   int    `json:"id"`
}

type User struct {
	Email string `json:"email"`
	Id    int    `json:"id"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}
type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

func (db *DB) CreateUser(email string) (User, error) {
	user := User{}
	dbs, err := db.load()
	if err != nil {
		return user, err
	}

	// NOTE: This should be valid as long as we only modify the database here, but might need to be made more robust later if we start deleting users
	id := len(dbs.Users) + 1

	user.Email = email
	user.Id = id

	dbs.Users[id] = user

	err = db.write(dbs)
	if err != nil {
		log.Printf("Error writing database when adding user: %v", user)
	}
	return user, err
}

func (db *DB) GetSortedUsers() ([]User, error) {
	dbs, err := db.load()
	if err != nil {
		log.Printf("Error loading database while getting users: %v", err)
		return nil, err
	}

	users := make([]User, len(dbs.Users))
	i := 0
	for k := range dbs.Users {
		users[i] = dbs.Users[k]
		i++
	}

	slices.SortFunc(users, func(a, b User) int {
		return a.Id - b.Id
	})
	return users, nil
}

func (db *DB) GetUser(id int) (User, error) {
	dbs, err := db.load()
	if err != nil {
		return User{}, err
	}

	user, exists := dbs.Users[id]
	if !exists {
		return User{}, errors.New("User with requested id doesn't exist")
	}

	return user, nil
}
func (db *DB) CreateChirp(body string) (Chirp, error) {
	chirp := Chirp{}
	dbs, err := db.load()
	if err != nil {
		return chirp, err
	}

	// NOTE: This should be valid as long as we only modify the database here, but might need to be made more robust later if we start deleting chirps
	id := len(dbs.Chirps) + 1

	chirp.Body = body
	chirp.Id = id

	dbs.Chirps[id] = chirp

	err = db.write(dbs)
	if err != nil {
		log.Printf("Error writing database when adding chirp: %v", chirp)
	}
	return chirp, err
}

func (db *DB) GetSortedChirps() ([]Chirp, error) {
	dbs, err := db.load()
	if err != nil {
		log.Printf("Error loading database while getting chirps: %v", err)
		return nil, err
	}

	chirps := make([]Chirp, len(dbs.Chirps))
	i := 0
	for k := range dbs.Chirps {
		chirps[i] = dbs.Chirps[k]
		i++
	}

	slices.SortFunc(chirps, func(a, b Chirp) int {
		return a.Id - b.Id
	})
	return chirps, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	dbs, err := db.load()
	if err != nil {
		return Chirp{}, err
	}

	chirp, exists := dbs.Chirps[id]
	if !exists {
		return Chirp{}, errors.New("Chirp with requested id doesn't exist")
	}

	return chirp, nil
}

func NewDB(path string) (*DB, error) {
	log.Println("Creating new database connection")
	db := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := db.ensure()
	if err != nil {
		return nil, err
	}

	return &db, nil
}

func (db *DB) ensure() error {
	log.Printf("Ensure that database at %v exists", db.path)
	dbs := DBStructure{Chirps: make(map[int]Chirp), Users: make(map[int]User)}
	_, err := os.ReadFile(db.path)
	if err != nil {
		err = db.write(dbs)
	}

	return err
}

func (db *DB) load() (DBStructure, error) {
	dbs := DBStructure{}
	db.mux.RLock()
	defer db.mux.RUnlock()
	data, err := os.ReadFile(db.path)
	if err != nil {
		log.Printf("Error reading database file %v while loading: %v", db.path, err)
		return dbs, err
	}

	err = json.Unmarshal(data, &dbs)
	if err != nil {
		log.Printf("Error parsing JSON read from database: %v\ndata: %v", err, data)
		return dbs, err
	}
	return dbs, nil
}

func (db *DB) write(dbs DBStructure) error {
	log.Println("Writing to database at", db.path)
	data, err := json.Marshal(dbs)
	if err != nil {
		return err
	}

	db.mux.Lock()
	defer db.mux.Unlock()
	err = os.WriteFile(db.path, data, 0600)
	if err != nil {
		log.Println("Error writing to database:", err)
	}
	return err
}
