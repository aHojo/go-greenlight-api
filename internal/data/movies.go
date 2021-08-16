package data

import (
	"database/sql"
	"time"

	"github.com/ahojo/greenlight/internal/validator"
)

// Models wraps all of our database models
type Models struct {
	Movies MovieModel
}

// Creates a Models that holds all of our database models. 
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
// Movie data that we will reutrn as JSON
// The props all need to be exported
type Movie struct {
	ID        int64     `json:"id"` // Unique int ID for the movie
	CreatedAt time.Time `json:"-"`  // Timestamp for when the movie is added to our db - not relevant so "-" means to never show it.
	Title     string    `json:"title"`
	Year      int32     `json:"int32,omitempty"` // Release year
	// The Runtime MarshalJSON() receiver will be called now.
	Runtime Runtime `json:"runtime,omitempty"` // omitempty means to not show it if there is no data.
	// If you want to use omitempty and not change the key name then you can leave it blank in the struct tag — like this: json:",omitempty". Notice that the leading comma is still required.
	Genres  []string `json:"genres,omitempty"`
	Version int32    `json:"version"` // incremented everytime the movie info is updated
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	// Check() method to execute the validation checks. Adds the provided key and error message to the errors map.
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
}

// MovieModel wraps our db connection
type MovieModel struct {
	DB *sql.DB
}

// Insert inserts a new movie record
func (m *MovieModel) Insert(movie *Movie) error {
	return nil 
}
// Get gets a specific movie from our database
func (m *MovieModel) Get(movie *Movie) error {
	return nil 
}

// Update updates a specific movie from our database
func (m *MovieModel) Update(movie *Movie) error {
	return nil 
}

// Delete
func (m *MovieModel) Delete(id int64) error {
	return nil 
}

