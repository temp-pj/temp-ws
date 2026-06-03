package music

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store { db: db }
}

func (s *Store) FetchSongs(ctx context.Context, c Category, count int, timeLimit int) ([]Song, error) {
	query := `
		SELECT r.mbid, r.name, r.isrc, a.name AS artist_name, r.release_date, r.duration
		FROM recordings r
		JOIN artists a ON r.artist_mbid = a.mbid
		JOIN artist_tags at ON a.mbid = at.artist_mbid
		WHERE at.tag = $1
		  AND EXTRACT(YEAR FROM r.release_date) >= $2
		  AND EXTRACT(YEAR FROM r.release_date) <= $3
		  AND a.area = $4
		  AND a.type = $5
		  AND ($6 = '' OR a.gender = $6 OR a.gender IS NULL)
		  AND r.listen_count >= 1000
		  AND r.duration >= ($7 + 10) * 1000
		ORDER BY RANDOM()
		LIMIT $8;
	`

	rows, err := s.db.QueryContext(
		ctx,
		query,
		c.Genre,
		c.StartYear,
		c.EndYear,
		c.Country,
		c.ArtistType,
		c.Gender,
		timeLimit,
		count,
	)

	if err != nil {
		return nil, fmt.Errorf("DB에서 곡 가져오다가 에러남: %w", err)
	}

	defer func() { _ = rows.Close() }() 

	var songs []Song

	for rows.Next() {
		var song Song
		var releaseDate sql.NullTime

		err := rows.Scan(
			&song.MBID,
			&song.Title,
			&song.ISRC,
			&song.Artist,
			&releaseDate,
			&song.Duration,
		)

		if err != nil {
			return nil, fmt.Errorf("Song 저장 도중 실패: %w", err)
		}

		if releaseDate.Valid {
			song.ReleaseDate = releaseDate.Time
		}

		songs = append(songs, song)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("데이터에 에러 발생: %w", err)
	}

	return songs, nil
}