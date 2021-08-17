package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/ahojo/greenlight/internal/validator"
	"github.com/lib/pq"
)

var (
	ErrRecordNotFound = errors.New("no record matching request")
	ErrEditConflict = errors.New("edit conflict")
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
	// Define the SQL query for inserting a new record in the movies table and returning the system generated data
	query := `INSERT INTO movies (title, year, runtime, genres) 
						VALUES ($1, $2, $3, $4)
						RETURNING id, created_at, version`

	// Create an args slice containing the values for the placeholder parameters from the movie struct
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Execute the query.
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get gets a specific movie from our database
func (m *MovieModel) Get(id int64) (*Movie, error) {
	// The postgresql bigserial type starts autoincrementing at 1.
	// No movies will have a value below 1.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Sql query
	stmt := `SELECT id,created_at,title,year,runtime,genres,version
					 FROM movies
					 WHERE id = $1`
	// declare a movie
	var movie Movie

	// Execute the query NOTE: that we have to use pg.Array() here
	err := m.DB.QueryRow(stmt, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, err
}

// Update updates a specific movie from our database
func (m *MovieModel) Update(movie *Movie) error {

	// Add version = $6, so we can stop race conditions
	query := `
	UPDATE movies
	SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version 
	`

	// create the arg slice contaninig the values for the placeholder params.
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version, // Add the expected movie version
	} 

	// If no matching row could be found (version has been changed) 
	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete
func (m *MovieModel) Delete(id int64) error {
	// ids can't be less than 1
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE FROM movies
	WHERE id = $1`

	// Returns sql.Result for how many rows affected
	result,err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() to get the # of rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}


