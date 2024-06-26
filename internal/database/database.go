package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Chirp struct {
	Body     string `json:"body"`
	Id       int    `json:"id"`
	AuthorId int    `json:"author_id"`
}

type user struct {
	Email       string `json:"email"`
	Hash        []byte `json:"hash"`
	Id          int    `json:"id"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

type SafeUser struct {
	Email       string `json:"email"`
	Id          int    `json:"id"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

func (u user) clean() SafeUser {
	return SafeUser{Email: u.Email, Id: u.Id, IsChirpyRed: u.IsChirpyRed}
}

type DB struct {
	path string
	mux  *sync.RWMutex
}
type DBStructure struct {
	Chirps        map[int]Chirp        `json:"chirps"`
	Users         map[int]user         `json:"users"`
	RevokedTokens map[string]time.Time `json:"revoked_tokens"`
	// Cheap way to get unique ids
	NextChirpId int `json:"nextChirpId"`
}

func (db *DB) writeUser(user user, password string, newUser bool) (SafeUser, error) {
	if !newUser && user.Id <= 0 {
		log.Fatal("Invalid operation: Overwrite user with negative id:", user)
	}
	if newUser && user.Id != 0 {
		log.Fatal("Invalid operation: Tried to specify the id when creating new user:", user)
	}

	dbs, err := db.load()
	if err != nil {
		return user.clean(), err
	}
	if newUser {
		user.Id = (len(dbs.Users) + 1)
	}

	// NOTE: This should be valid as long as we only modify the database here, but might need to be made more robust later if we start deleting users

	user.Hash, err = bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		log.Printf("Error hashing password when writing user: %v", user)
	}

	dbs.Users[user.Id] = user

	err = db.write(dbs)
	if err != nil {
		log.Printf("Error writing database when writing user: %v", user)
	}
	return user.clean(), err
}

func (db *DB) CreateUser(email, password string) (SafeUser, error) {
	_, err := db.getUserByEmail(email)
	if err == nil {
		return SafeUser{}, errors.New("User with given email already exists")
	}
	user := user{Email: email}
	return db.writeUser(user, password, true)
}

func (db *DB) UpdateUser(id int, email, password string) (SafeUser, error) {
	suser, err := db.GetUser(id)
	if err == nil {
		return SafeUser{}, errors.New("User with given id not found")
	}
	user := user{Email: email, Id: id, IsChirpyRed: suser.IsChirpyRed}
	return db.writeUser(user, password, false)
}

func (db *DB) UpgradeUser(id int) error {
	dbs, err := db.load()
	if err != nil {
		return err
	}

	user, exists := dbs.Users[id]
	if !exists {
		return fmt.Errorf("User doesn't exist")
	}

	user.IsChirpyRed = true
	// NOTE: You can't update map values, only reassign them. So either we rewrite entries every time, or use maps of pointers.
	dbs.Users[id] = user
	db.write(dbs)
	return nil
}

func (db *DB) GetSortedUsers() ([]SafeUser, error) {
	dbs, err := db.load()
	if err != nil {
		log.Printf("Error loading database while getting users: %v", err)
		return nil, err
	}

	users := make([]SafeUser, len(dbs.Users))
	i := 0
	for k := range dbs.Users {
		users[i] = dbs.Users[k].clean()
		i++
	}

	slices.SortFunc(users, func(a, b SafeUser) int {
		return a.Id - b.Id
	})
	return users, nil
}

func (db *DB) GetUser(id int) (SafeUser, error) {
	dbs, err := db.load()
	if err != nil {
		return SafeUser{}, err
	}

	user, exists := dbs.Users[id]
	if !exists {
		return SafeUser{}, errors.New("User with requested id doesn't exist")
	}

	return user.clean(), nil
}

func (db *DB) getUserByEmail(email string) (user, error) {
	dbs, err := db.load()
	if err != nil {
		return user{}, err
	}
	for _, v := range dbs.Users {
		if v.Email == email {
			return v, nil
		}
	}
	return user{}, errors.New("User with requested email doesn't exist")
}

func (db *DB) ValidateLogin(email, password string) (SafeUser, error) {
	safeUser := SafeUser{}
	user, err := db.getUserByEmail(email)
	if err != nil {
		return safeUser, errors.New("User email not found")
	}

	err = bcrypt.CompareHashAndPassword(user.Hash, []byte(password))
	if err == nil {
		safeUser = user.clean()
	}
	return safeUser, err
}

func (db *DB) CreateChirp(body string, authorId int) (Chirp, error) {
	chirp := Chirp{}
	dbs, err := db.load()
	if err != nil {
		return chirp, err
	}

	id := dbs.NextChirpId
	dbs.NextChirpId++

	chirp.Body = body
	chirp.Id = id
	chirp.AuthorId = authorId

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

func (db *DB) DeleteChirp(id int) error {
	dbs, err := db.load()
	if err != nil {
		return err
	}
	delete(dbs.Chirps, id)
	return nil
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
	dbs := DBStructure{Chirps: make(map[int]Chirp), Users: make(map[int]user), RevokedTokens: make(map[string]time.Time), NextChirpId: 1}
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

func (db *DB) TokenRevoked(tokenString string) (bool, error) {
	dbs, err := db.load()
	if err != nil {
		return false, err
	}
	_, ok := dbs.RevokedTokens[tokenString]
	return ok, nil
}

func (db *DB) RevokeToken(tokenString string, time time.Time) error {
	dbs, err := db.load()
	if err != nil {
		return err
	}
	dbs.RevokedTokens[tokenString] = time
	err = db.write(dbs)
	return err
}
