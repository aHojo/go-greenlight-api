package data

import "time"

// Movie data that we will reutrn as JSON
// The props all need to be exported
type Movie struct {
	ID        int64     `json:"id"`         // Unique int ID for the movie
	CreatedAt time.Time `json:"-"` // Timestamp for when the movie is added to our db - not relevant so "-" means to never show it.
	Title     string    `json:"title"`
	Year      int32     `json:"int32,omitempty"` // Release year
	// The Runtime MarshalJSON() receiver will be called now. 
	Runtime   Runtime     `json:"runtime,omitempty"` // omitempty means to not show it if there is no data.
	// If you want to use omitempty and not change the key name then you can leave it blank in the struct tag — like this: json:",omitempty". Notice that the leading comma is still required.
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"` // incremented everytime the movie info is updated
}
