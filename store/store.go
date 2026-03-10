package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"myproject/models"
)

var ErrNotFound = errors.New("not found")
var ErrAccessDenied = errors.New("access denied")

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// User operations
func (s *Store) CreateUser(email, passwordHash, name string) (*models.User, error) {
	return s.CreateUserWithRole(email, passwordHash, name, "user")
}

func (s *Store) CreateUserWithRole(email, passwordHash, name, role string) (*models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		`INSERT INTO users (email, password_hash, name, role) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, email, name, role, created_at, updated_at`,
		email, passwordHash, name, role,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, name, role, created_at, updated_at 
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &u, err
}

func (s *Store) GetUserByID(id int) (*models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, name, role, created_at, updated_at 
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &u, err
}

func (s *Store) ListUsers() ([]models.User, error) {
	rows, err := s.db.Query(
		`SELECT id, email, name, role, created_at, updated_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
		users = append(users, u)
	}
	return users, nil
}

func (s *Store) UpdateUserRole(id int, role string) error {
	_, err := s.db.Exec(`UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`, role, id)
	return err
}

func (s *Store) DeleteUser(id int) error {
	_, err := s.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}

// Movie operations
func (s *Store) CreateMovie(m *models.Movie) error {
	castJSON, _ := json.Marshal(m.Cast)
	genresJSON, _ := json.Marshal(m.Genres)
	tagsJSON, _ := json.Marshal(m.Tags)

	return s.db.QueryRow(
		`INSERT INTO movies (title, description, file_path, thumbnail_path, duration, owner_id, content_type, parent_id, season_number, episode_number, cast_members, director, release_year, genres, tags)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id, created_at, updated_at`,
		m.Title, m.Description, m.FilePath, m.ThumbnailPath, m.Duration, m.OwnerID, m.ContentType, m.ParentID, m.SeasonNum, m.EpisodeNum, castJSON, m.Director, m.ReleaseYear, genresJSON, tagsJSON,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

func (s *Store) GetMovie(id int) (*models.Movie, error) {
	var m models.Movie
	var castJSON, genresJSON, tagsJSON []byte
	err := s.db.QueryRow(
		`SELECT id, title, description, 
			COALESCE(file_path, ''), COALESCE(thumbnail_path, ''), COALESCE(duration, 0), COALESCE(owner_id, 0),
			COALESCE(content_type, 'movie'), parent_id, season_number, episode_number,
			COALESCE(cast_members, '[]'), COALESCE(director, ''), COALESCE(release_year, 0),
			COALESCE(genres, '[]'), COALESCE(tags, '[]'), COALESCE(mood, ''), created_at, updated_at
		 FROM movies WHERE id = $1`, id,
	).Scan(&m.ID, &m.Title, &m.Description, &m.FilePath, &m.ThumbnailPath, &m.Duration, &m.OwnerID,
		&m.ContentType, &m.ParentID, &m.SeasonNum, &m.EpisodeNum, &castJSON, &m.Director, &m.ReleaseYear,
		&genresJSON, &tagsJSON, &m.Mood, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(castJSON) > 0 {
		json.Unmarshal(castJSON, &m.Cast)
	}
	if len(genresJSON) > 0 {
		json.Unmarshal(genresJSON, &m.Genres)
	}
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &m.Tags)
	}
	if m.Cast == nil {
		m.Cast = []string{}
	}
	if m.Genres == nil {
		m.Genres = []string{}
	}
	if m.Tags == nil {
		m.Tags = []string{}
	}
	return &m, nil
}

func (s *Store) ListMoviesForUser(userID int) ([]models.Movie, error) {
	// Only return root content (movies and series) in Browse, not individual episodes
	rows, err := s.db.Query(`
		SELECT id, title, description, 
			COALESCE(file_path, ''), COALESCE(thumbnail_path, ''), COALESCE(duration, 0), COALESCE(owner_id, 0),
			COALESCE(content_type, 'movie'), parent_id, season_number, episode_number,
			COALESCE(cast_members, '[]'), COALESCE(director, ''), COALESCE(release_year, 0),
			COALESCE(genres, '[]'), COALESCE(tags, '[]'), COALESCE(mood, ''), created_at, updated_at
		FROM movies 
		WHERE COALESCE(content_type, 'movie') != 'episode'
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		var castJSON, genresJSON, tagsJSON []byte
		rows.Scan(&m.ID, &m.Title, &m.Description, &m.FilePath, &m.ThumbnailPath,
			&m.Duration, &m.OwnerID, &m.ContentType, &m.ParentID, &m.SeasonNum, &m.EpisodeNum, &castJSON, &m.Director, &m.ReleaseYear, &genresJSON, &tagsJSON, &m.Mood, &m.CreatedAt, &m.UpdatedAt)
		if len(castJSON) > 0 {
			json.Unmarshal(castJSON, &m.Cast)
		}
		if len(genresJSON) > 0 {
			json.Unmarshal(genresJSON, &m.Genres)
		}
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &m.Tags)
		}
		if m.Cast == nil {
			m.Cast = []string{}
		}
		if m.Genres == nil {
			m.Genres = []string{}
		}
		if m.Tags == nil {
			m.Tags = []string{}
		}
		movies = append(movies, m)
	}
	return movies, nil
}

func (s *Store) SearchMovies(userID int, query string) ([]models.Movie, error) {
	rows, err := s.db.Query(`
		SELECT id, title, description, 
			COALESCE(file_path, ''), COALESCE(thumbnail_path, ''), COALESCE(duration, 0), COALESCE(owner_id, 0),
			COALESCE(content_type, 'movie'), parent_id, season_number, episode_number,
			COALESCE(cast_members, '[]'), COALESCE(director, ''), COALESCE(release_year, 0),
			COALESCE(genres, '[]'), COALESCE(tags, '[]'), COALESCE(mood, ''), created_at, updated_at
		FROM movies 
		WHERE title ILIKE $1 OR description ILIKE $1 OR COALESCE(director, '') ILIKE $1 OR COALESCE(cast_members, '[]')::text ILIKE $1 OR COALESCE(genres, '[]')::text ILIKE $1
		ORDER BY created_at DESC`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		var castJSON, genresJSON, tagsJSON []byte
		rows.Scan(&m.ID, &m.Title, &m.Description, &m.FilePath, &m.ThumbnailPath,
			&m.Duration, &m.OwnerID, &m.ContentType, &m.ParentID, &m.SeasonNum, &m.EpisodeNum, &castJSON, &m.Director, &m.ReleaseYear, &genresJSON, &tagsJSON, &m.Mood, &m.CreatedAt, &m.UpdatedAt)
		if len(castJSON) > 0 {
			json.Unmarshal(castJSON, &m.Cast)
		}
		if len(genresJSON) > 0 {
			json.Unmarshal(genresJSON, &m.Genres)
		}
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &m.Tags)
		}
		if m.Cast == nil {
			m.Cast = []string{}
		}
		if m.Genres == nil {
			m.Genres = []string{}
		}
		if m.Tags == nil {
			m.Tags = []string{}
		}
		movies = append(movies, m)
	}
	return movies, nil
}

func (s *Store) GetEpisodesForSeries(seriesID int) ([]models.Movie, error) {
	rows, err := s.db.Query(`
		SELECT id, title, description, 
			COALESCE(file_path, ''), COALESCE(thumbnail_path, ''), COALESCE(duration, 0), COALESCE(owner_id, 0),
			COALESCE(content_type, 'episode'), parent_id, season_number, episode_number,
			COALESCE(cast_members, '[]'), COALESCE(director, ''), COALESCE(release_year, 0),
			COALESCE(genres, '[]'), COALESCE(tags, '[]'), COALESCE(mood, ''), created_at, updated_at
		FROM movies 
		WHERE parent_id = $1 AND COALESCE(content_type, 'movie') = 'episode'
		ORDER BY season_number ASC, episode_number ASC`, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var episodes []models.Movie
	for rows.Next() {
		var m models.Movie
		var castJSON, genresJSON, tagsJSON []byte
		rows.Scan(&m.ID, &m.Title, &m.Description, &m.FilePath, &m.ThumbnailPath,
			&m.Duration, &m.OwnerID, &m.ContentType, &m.ParentID, &m.SeasonNum, &m.EpisodeNum, &castJSON, &m.Director, &m.ReleaseYear, &genresJSON, &tagsJSON, &m.Mood, &m.CreatedAt, &m.UpdatedAt)
		if len(castJSON) > 0 {
			json.Unmarshal(castJSON, &m.Cast)
		}
		if len(genresJSON) > 0 {
			json.Unmarshal(genresJSON, &m.Genres)
		}
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &m.Tags)
		}
		if m.Cast == nil {
			m.Cast = []string{}
		}
		if m.Genres == nil {
			m.Genres = []string{}
		}
		if m.Tags == nil {
			m.Tags = []string{}
		}
		episodes = append(episodes, m)
	}
	return episodes, nil
}

func (s *Store) UpdateMovie(m *models.Movie) error {
	_, err := s.db.Exec(`
		UPDATE movies SET title=$1, description=$2, updated_at=NOW() 
		WHERE id=$3`, m.Title, m.Description, m.ID)
	return err
}

func (s *Store) UpdateMovieFull(m *models.Movie) error {
	castJSON, _ := json.Marshal(m.Cast)
	genresJSON, _ := json.Marshal(m.Genres)
	_, err := s.db.Exec(`
		UPDATE movies SET title=$1, description=$2, genres=$3, cast_members=$4, director=$5, release_year=$6, updated_at=NOW() 
		WHERE id=$7`,
		m.Title, m.Description, genresJSON, castJSON, m.Director, m.ReleaseYear, m.ID)
	return err
}

func (s *Store) DeleteMovie(id int) error {
	_, err := s.db.Exec(`DELETE FROM movies WHERE id = $1`, id)
	return err
}

// Watch history
func (s *Store) UpdateWatchProgress(userID, movieID, progress int) error {
	_, err := s.db.Exec(`
		INSERT INTO watch_history (user_id, movie_id, progress, last_watched)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, movie_id) DO UPDATE SET progress = $3, last_watched = NOW()`,
		userID, movieID, progress)
	return err
}

func (s *Store) GetWatchHistory(userID, limit int) ([]models.WatchHistory, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, movie_id, progress, last_watched 
		FROM watch_history WHERE user_id = $1 
		ORDER BY last_watched DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.WatchHistory
	for rows.Next() {
		var h models.WatchHistory
		rows.Scan(&h.ID, &h.UserID, &h.MovieID, &h.Progress, &h.LastWatched)
		history = append(history, h)
	}
	return history, nil
}

// Embedding operations
func (s *Store) UpdateMovieEmbedding(movieID int, embedding []float64) error {
	embJSON, _ := json.Marshal(embedding)
	_, err := s.db.Exec(`UPDATE movies SET embedding = $1 WHERE id = $2`, embJSON, movieID)
	return err
}

func (s *Store) FindSimilarMovies(excludeID int, embedding []float64, limit int) ([]models.Movie, error) {
	// For production, use pgvector extension for efficient similarity search
	// This is a simplified version that loads all movies
	rows, err := s.db.Query(`
		SELECT id, title, description, file_path, thumbnail_path, duration, 
		       owner_id, genres, tags, mood, embedding, created_at, updated_at
		FROM movies WHERE id != $1 AND embedding IS NOT NULL`, excludeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scoredMovie struct {
		movie models.Movie
		score float64
	}
	var scored []scoredMovie

	for rows.Next() {
		var m models.Movie
		var genresJSON, tagsJSON, embJSON []byte
		rows.Scan(&m.ID, &m.Title, &m.Description, &m.FilePath, &m.ThumbnailPath,
			&m.Duration, &m.OwnerID, &genresJSON, &tagsJSON, &m.Mood, &embJSON, &m.CreatedAt, &m.UpdatedAt)
		json.Unmarshal(genresJSON, &m.Genres)
		json.Unmarshal(tagsJSON, &m.Tags)
		json.Unmarshal(embJSON, &m.Embedding)

		if m.Embedding != nil {
			score := cosineSimilarity(embedding, m.Embedding)
			scored = append(scored, scoredMovie{m, score})
		}
	}

	// Sort by similarity score
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	var result []models.Movie
	for i := 0; i < len(scored) && i < limit; i++ {
		result = append(result, scored[i].movie)
	}
	return result, nil
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}
