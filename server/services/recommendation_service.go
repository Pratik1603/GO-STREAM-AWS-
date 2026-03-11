package services

import (
	"fmt"
	"strings"

	"server/myproject/models"
	"server/myproject/store"
)

type RecommendationService struct {
	store *store.Store
}

func NewRecommendationService(store *store.Store) *RecommendationService {
	return &RecommendationService{

		store: store,
	}
}

// OpenAI request/response structures
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// Get recommendations based on user's watch history and preferences
func (rs *RecommendationService) GetRecommendations(userID int, limit int) ([]models.Movie, error) {

	history, err := rs.store.GetWatchHistory(userID, 20)
	if err != nil {
		return nil, err
	}

	allMovies, err := rs.store.ListMoviesForUser()
	if err != nil {
		return nil, err
	}

	watchedIDs := make(map[int]bool)
	var watchedMovies []models.Movie
	var watchedTitles []string

	for _, h := range history {
		watchedIDs[h.MovieID] = true
		if movie, _ := rs.store.GetMovie(h.MovieID); movie != nil {
			watchedMovies = append(watchedMovies, *movie)
			watchedTitles = append(watchedTitles, fmt.Sprintf("%s: %s", movie.Title, movie.Description))
		}
	}

	var unwatched []models.Movie
	var unwatchedInfo []string
	for _, m := range allMovies {
		if !watchedIDs[m.ID] {
			unwatched = append(unwatched, m)
			unwatchedInfo = append(unwatchedInfo, fmt.Sprintf("ID:%d|%s|%s|%s", m.ID, m.Title, m.Description, strings.Join(m.Genres, ",")))
		}
	}

	if len(unwatched) == 0 {
		return []models.Movie{}, nil
	}

	return rs.getRecommendationsBasic(watchedMovies, unwatched, limit)
}

// Get recommendations using genre and keyword matching from watch history
func (rs *RecommendationService) getRecommendationsBasic(watchedMovies []models.Movie, unwatchedMovies []models.Movie, limit int) ([]models.Movie, error) {
	if len(watchedMovies) == 0 {

		if len(unwatchedMovies) > limit {
			return unwatchedMovies[:limit], nil
		}
		return unwatchedMovies, nil
	}

	watchedGenreMap := make(map[string]int)
	var allWatchedKeywords []string

	for _, movie := range watchedMovies {

		for _, genre := range movie.Genres {
			watchedGenreMap[strings.ToLower(genre)]++
		}
		keywords := extractKeywords(movie.Description)
		allWatchedKeywords = append(allWatchedKeywords, keywords...)
	}

	keywordMap := make(map[string]int)
	for _, kw := range allWatchedKeywords {
		keywordMap[kw]++
	}

	type movieScore struct {
		movie models.Movie
		score float64
	}

	var scores []movieScore

	for _, m := range unwatchedMovies {
		score := 0.0

		genreMatches := 0
		for _, mGenre := range m.Genres {
			if count, exists := watchedGenreMap[strings.ToLower(mGenre)]; exists {
				genreMatches += count // Weight by how many watched movies had this genre
			}
		}
		if len(watchedGenreMap) > 0 {
			avgGenresWatched := 0
			for _, count := range watchedGenreMap {
				avgGenresWatched += count
			}
			score += (float64(genreMatches) / float64(avgGenresWatched)) * 70
		}
		mDescLower := strings.ToLower(m.Description)
		keywordMatches := 0
		for keyword := range keywordMap {
			if strings.Contains(mDescLower, keyword) {
				keywordMatches++
			}
		}
		if len(keywordMap) > 0 {
			score += (float64(keywordMatches) / float64(len(keywordMap))) * 20
		}

		diversityBonus := 0.0
		for _, genre := range m.Genres {
			if count, exists := watchedGenreMap[strings.ToLower(genre)]; !exists {
				diversityBonus += 5
			} else if count <= 1 {
				diversityBonus += 3
			}
		}
		if diversityBonus > 10 {
			diversityBonus = 10
		}
		score += diversityBonus

		if score > 0 {
			scores = append(scores, movieScore{movie: m, score: score})
		}
	}

	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	var result []models.Movie
	maxResults := limit
	if len(scores) < maxResults {
		maxResults = len(scores)
	}
	for i := 0; i < maxResults; i++ {
		result = append(result, scores[i].movie)
	}

	return result, nil
}

// Get similar movies using embeddings (with fallback for when API key is not available)
func (rs *RecommendationService) GetSimilarMovies(movieID int, limit int) ([]models.Movie, error) {
	movie, err := rs.store.GetMovie(movieID)
	if err != nil {
		return nil, err
	}
	return rs.getSimilarMoviesBasic(movie, limit)
}

// Get similar movies using genre and keyword matching (no API key required)
func (rs *RecommendationService) getSimilarMoviesBasic(movie *models.Movie, limit int) ([]models.Movie, error) {

	allMovies, err := rs.store.ListMoviesForUser()
	if err != nil {
		return nil, err
	}

	type movieScore struct {
		movie models.Movie
		score float64
	}

	var scores []movieScore
	descKeywords := extractKeywords(movie.Description)
	movieGenreMap := make(map[string]bool)
	for _, g := range movie.Genres {
		movieGenreMap[strings.ToLower(g)] = true
	}

	for _, m := range allMovies {
		if m.ID == movie.ID {
			continue
		}

		score := 0.0

		genreMatches := 0
		for _, mGenre := range m.Genres {
			if movieGenreMap[strings.ToLower(mGenre)] {
				genreMatches++
			}
		}
		if len(movie.Genres) > 0 {
			score += (float64(genreMatches) / float64(len(movie.Genres))) * 70
		}

		mDescLower := strings.ToLower(m.Description)
		keywordMatches := 0
		for _, keyword := range descKeywords {
			if strings.Contains(mDescLower, keyword) {
				keywordMatches++
			}
		}
		if len(descKeywords) > 0 {
			score += (float64(keywordMatches) / float64(len(descKeywords))) * 20
		}

		titleScore := calculateTitleSimilarity(movie.Title, m.Title)
		score += titleScore * 10

		if score > 0 {
			scores = append(scores, movieScore{movie: m, score: score})
		}
	}

	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	var result []models.Movie
	maxResults := limit
	if len(scores) < maxResults {
		maxResults = len(scores)
	}
	for i := 0; i < maxResults; i++ {
		result = append(result, scores[i].movie)
	}

	return result, nil
}

func extractKeywords(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	var keywords []string
	stopWords := map[string]bool{
		"and": true, "the": true, "a": true, "an": true, "or": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"with": true, "for": true, "of": true, "in": true, "on": true, "at": true,
	}

	for _, word := range words {

		word = strings.Trim(word, ".,!?;:")
		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	return keywords
}

// Calculate similarity between two movie titles
func calculateTitleSimilarity(title1, title2 string) float64 {
	words1 := strings.Fields(strings.ToLower(title1))
	words2 := strings.Fields(strings.ToLower(title2))

	matches := 0
	for _, w1 := range words1 {
		w1 = strings.Trim(w1, ".,!?;:")
		for _, w2 := range words2 {
			w2 = strings.Trim(w2, ".,!?;:")
			if w1 == w2 {
				matches++
			}
		}
	}

	maxLen := len(words1)
	if len(words2) > maxLen {
		maxLen = len(words2)
	}

	if maxLen == 0 {
		return 0
	}

	return float64(matches) / float64(maxLen)
}

func (rs *RecommendationService) UpdateWatchProgress(userID, movieID, progress int) error {
	return rs.store.UpdateWatchProgress(userID, movieID, progress)
}

func (rs *RecommendationService) GetWatchHistory(userID, limit int) ([]models.WatchHistory, error) {
	return rs.store.GetWatchHistory(userID, limit)
}
