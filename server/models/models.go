package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Movie struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	FilePath      string    `json:"file_path"`
	ThumbnailPath string    `json:"thumbnail_path"`
	Duration      int       `json:"duration"`
	OwnerID       int       `json:"owner_id"`
	VideoURL      string    `json:"video_url,omitempty"`
	ThumbnailURL  string    `json:"thumbnail_url,omitempty"`
	ContentType   string    `json:"content_type"`
	ParentID      *int      `json:"parent_id,omitempty"`
	SeasonNum     *int      `json:"season_number,omitempty"`
	EpisodeNum    *int      `json:"episode_number,omitempty"`
	Cast          []string  `json:"cast_members"`
	Director      string    `json:"director"`
	ReleaseYear   int       `json:"release_year"`
	Genres        []string  `json:"genres"`
	Tags          []string  `json:"tags"`
	Mood          string    `json:"mood"`
	Embedding     []float64 `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type MovieAccess struct {
	ID        int       `json:"id"`
	MovieID   int       `json:"movie_id"`
	UserID    int       `json:"user_id"`
	GrantedBy int       `json:"granted_by"`
	CreatedAt time.Time `json:"created_at"`
}

type WatchHistory struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	MovieID     int       `json:"movie_id"`
	Progress    int       `json:"progress"`
	LastWatched time.Time `json:"last_watched"`
}

type MovieAnalysis struct {
	Genres         []string `json:"genres"`
	Tags           []string `json:"tags"`
	Mood           string   `json:"mood"`
	Themes         []string `json:"themes"`
	TargetAudience string   `json:"target_audience"`
	SimilarTo      []string `json:"similar_to"`
}
