package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ahojo/greenlight/internal/validator"
	"github.com/lib/pq"
)

var (
	ErrRecordNotFound = errors.New("no record matching request")
	ErrEditConflict   = errors.New("edit conflict")
)

// Models wraps all of our database models
type Models struct {
	Movies MovieModel
	Users  UserModel
}

// Creates a Models that holds all of our database models.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users: UserModel{DB: db},
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

/* DATABASE QUERIES */

/**
GET METHODS
*/
// Get gets a specific movie from our database
func (m *MovieModel) Get(id int64) (*Movie, error) {
	// The postgresql bigserial type starts autoincrementing at 1.
	// No movies will have a value below 1.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Sql query
	// pq_sleep(10) to simulate a long running query
	// stmt := `SELECT pg_sleep(10),id,created_at,title,year,runtime,genres,version
	// 				 FROM movies
	// 				 WHERE id = $1`
	stmt := `SELECT id,created_at,title,year,runtime,genres,version
					 FROM movies
					 WHERE id = $1`
	// declare a movie
	var movie Movie

	// ctx.WithTimeout() funciton to carry a 3 second timeout deadline.
	// emtpy context.Background() is the parent context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// IMPORTANT, use defer cancel() so we can cancel the context before Get() returns.
	defer cancel()

	// Execute the query NOTE: that we have to use pg.Array() here
	// err := m.DB.QueryRow(stmt, id).Scan(
	// 	&[]byte{}, // for the pg_sleep(10)
	// 	&movie.ID,
	// 	&movie.CreatedAt,
	// 	&movie.Title,
	// 	&movie.Year,
	// 	&movie.Runtime,
	// 	pq.Array(&movie.Genres),
	// 	&movie.Version,
	// )
	err := m.DB.QueryRowContext(ctx, stmt, id).Scan(
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

// GetAll returns a slice of Movies.
func (m *MovieModel) GetAll(title string, gernres []string, filters Filters) ([]*Movie, Metadata, error) {
	// Sql Query
	// query := `
	// SELECT id, created_at, title, year, runtime, genres, version
	// FROM movies
	// ORDER BY id
	// `

	/*
			This SQL query is designed so that each of the filters behaves like it is ‘optional’. For
		example, the condition (LOWER(title) = LOWER($1) OR $1 = '')  will evaluate as true if
		the placeholder parameter $1 is a case-insensitive match for the movie title or the
		placeholder parameter equals ''. So this filter condition will essentially be ‘skipped’ when
		movie title being searched for is the empty string "".

			The (genres @> $2 OR $2 = '{}')  condition works in the same way. The  @> symbol is the
		‘contains’ operator for PostgreSQL arrays, and this condition will return true if all values in
		the placeholder parameter $2 are contained in the database  genres field or the placeholder
		parameter contains an empty array.

		https://www.postgresql.org/docs/9.6/functions-array.html
	*/
	// query := `
	// SELECT id, created_at, title, year, runtime, genres, version
	// FROM movies
	// WHERE (LOWER(title) = LOWER($1) OR $1 = '')
	// AND (genres @> $2 OR $2 = '{}')
	// ORDER BY id`

	/* Add FULL TEXT SEARCH PostgreSQL feature
	The to_tsvector('simple', title) function takes a movie title and splits it into lexemes.
	We specify the simple configuration, which means that the lexemes are just lowercase
	versions of the words in the title

	The plainto_tsquery('simple', $1) function takes a search value and turns it into a
	formatted query term.

	It normalizes the search value (again using the simple configuration), strips any special characters, and
	inserts the and operator & between the words.
	The @@ operator is the matches operator. In our statement we are using it to check whether
	the generated query term matches the lexemes.
	*/
	// query := `
	// SELECT id, created_at, title, year, runtime, genres, version
	// FROM movies
	// WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	// AND (genres @> $2 OR $2 = '{}')
	// ORDER BY id
	// `

	/* could have also used ILIKE
		SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE (title ILIKE $1 OR $1 = '')
	AND (genres @> $2 OR $2 = '{}')
	ORDER BY id
	*/

	// Add an ORDER BY clause and interpolate the sort column and direction. Importantly
	// notice that we also include a secondary sort on the movie ID to ensure a
	// consistent ordering.
	// Added the window function to count the number of (filtered) records
	query := fmt.Sprintf(`
	SELECT COUNT(*) OVER(),id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') 
	AND (genres @> $2 OR $2 = '{}')     
	ORDER BY %s %s, id ASC
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	// 3 second context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{
		title,
		pq.Array(gernres),
		filters.limit(),
		filters.offset(),
	}
	// Get back the data from the database. Cancels if takes too long
	// Title and genres have the default params.
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	// Make sure to close the rows stream return
	defer rows.Close()

	totalRecords := 0
	// data structure to hold all of our movies
	var movies = []*Movie{}

	// Iterate through the rows returned
	for rows.Next() {
		var movie Movie

		// Scan the values from the row into the Movie
		// Note: pq.Array() again
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into total records
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		movies = append(movies, &movie)
	}

	// When the rows.Next() finishes, if there is an error it's in rows.Err()
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Generate a Metadata struct, passing in the total record count and pagination params from the client
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil

}

// Insert inserts a new movie record
func (m *MovieModel) Insert(movie *Movie) error {
	// Define the SQL query for inserting a new record in the movies table and returning the system generated data
	query := `INSERT INTO movies (title, year, runtime, genres) 
						VALUES ($1, $2, $3, $4)
						RETURNING id, created_at, version`

	// Create an args slice containing the values for the placeholder parameters from the movie struct
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query.
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Update updates a specific movie from our database
func (m *MovieModel) Update(movie *Movie) error {

	/* potential to use uuid here
	UPDATE movies
	SET title = $1, year = $2, runtime = $3, genres = $4, version = uuid_generate_v4()
	WHERE id = $5 AND
	**/
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// If no matching row could be found (version has been changed)
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Returns sql.Result for how many rows affected
	result, err := m.DB.ExecContext(ctx, query, id)
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
