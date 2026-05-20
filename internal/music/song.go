package music


type Song struct {
	Title string		`json:"title"`
	Artist string		`json:"artist"`
	PreviewURL string	`json:"previewURL"`
	WrappedID string	`json:"wrappedID"`
}

func (s Song) WrappedURL()string {
	/// TO-DO: 프록시 구현 예정 "https://audio.apple.com/preview/park-hyoshin-wildflower.m4a" -> "https://서버.com/audio/a3f8x9k2"
	return s.PreviewURL
}