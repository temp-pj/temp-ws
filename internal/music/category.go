package music

type Category struct {
	StartYear int		`json:"startYear"`
	EndYear int			`json:"endYear"`
	ArtistType string	`json:"artistType"`
	Country string		`json:"country"`
	Gender string		`json:"gender"`
	Genre string		`json:"genre"`
}