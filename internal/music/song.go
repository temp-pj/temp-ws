package music

import "time"


type Song struct {
	MBID string				`json:"mbid"`
	Title string			`json:"title"`
	ISRC string				`json:"isrc"`
	Artist string			`json:"artist"`
	ReleaseDate time.Time	`json:"release_date"`
	Duration int			`json:"duration"`
}
