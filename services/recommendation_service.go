package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"myproject/models"
	"myproject/store"
)

type RecommendationService struct {
	apiKey  string
	baseURL string
	store   *store.Store
}

func NewRecommendationService(store *store.Store) *RecommendationService {
	return &RecommendationService{
		apiKey:  os.Getenv("OPENAI_API_KEY"),
		baseURL: "https://api.openai.com/v1",
		store:   store,
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
	// Get user's watch history
	history, err := rs.store.GetWatchHistory(userID, 20)
	if err != nil {
		return nil, err
	}

	// Get all available movies for the user
	allMovies, err := rs.store.ListMoviesForUser(userID)
	if err != nil {
		return nil, err
	}

	// Filter out already watched movies
	watchedIDs := make(map[int]bool)
	var watchedTitles []string
	for _, h := range history {
		watchedIDs[h.MovieID] = true
		if movie, _ := rs.store.GetMovie(h.MovieID); movie != nil {
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

	// Use OpenAI to rank recommendations
	recommendedIDs, err := rs.getAIRecommendations(watchedTitles, unwatchedInfo, limit)
	if err != nil {
		// Fallback: return recent unwatched movies
		if len(unwatched) > limit {
			return unwatched[:limit], nil
		}
		return unwatched, nil
	}

	// Map recommended IDs to movies
	var recommended []models.Movie
	idToMovie := make(map[int]models.Movie)
	for _, m := range unwatched {
		idToMovie[m.ID] = m
	}
	for _, id := range recommendedIDs {
		if m, ok := idToMovie[id]; ok {
			recommended = append(recommended, m)
		}
	}

	return recommended, nil
}

func (rs *RecommendationService) getAIRecommendations(watched, available []string, limit int) ([]int, error) {
	if rs.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	prompt := fmt.Sprintf(`You are a movie recommendation system. Based on the user's watch history, recommend the best movies from the available list.

USER'S WATCH HISTORY:
%s

AVAILABLE MOVIES (format: ID|Title|Description|Genres):
%s

Return ONLY a JSON array of movie IDs in order of relevance, maximum %d recommendations.
Example: [5, 12, 3, 8]

Consider:
- Genre preferences based on watch history
- Similar themes and storylines
- Diversity in recommendations
- Quality and popularity signals from descriptions`,
		strings.Join(watched, "\n"),
		strings.Join(available, "\n"),
		limit)

	req := ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a movie recommendation AI. Respond only with valid JSON arrays of integers."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
		MaxTokens:   200,
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", rs.baseURL+"/chat/completions", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+rs.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("OpenAI error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the JSON array of IDs
	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	content = strings.Trim(content, "```json\n")
	content = strings.Trim(content, "```")

	var ids []int
	if err := json.Unmarshal([]byte(content), &ids); err != nil {
		return nil, err
	}

	return ids, nil
}

// Generate movie embedding for similarity search
func (rs *RecommendationService) GenerateEmbedding(text string) ([]float64, error) {
	if rs.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	req := EmbeddingRequest{
		Model: "text-embedding-3-small",
		Input: text,
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", rs.baseURL+"/embeddings", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+rs.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var embResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, err
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return embResp.Data[0].Embedding, nil
}

// Get similar movies using embeddings
func (rs *RecommendationService) GetSimilarMovies(movieID int, limit int) ([]models.Movie, error) {
	movie, err := rs.store.GetMovie(movieID)
	if err != nil {
		return nil, err
	}

	// Check if embedding exists, generate if not
	if movie.Embedding == nil {
		text := fmt.Sprintf("%s. %s. Genres: %s", movie.Title, movie.Description, strings.Join(movie.Genres, ", "))
		embedding, err := rs.GenerateEmbedding(text)
		if err != nil {
			return nil, err
		}
		rs.store.UpdateMovieEmbedding(movieID, embedding)
		movie.Embedding = embedding
	}

	// Find similar movies using cosine similarity
	return rs.store.FindSimilarMovies(movieID, movie.Embedding, limit)
}

func (rs *RecommendationService) UpdateWatchProgress(userID, movieID, progress int) error {
	return rs.store.UpdateWatchProgress(userID, movieID, progress)
}

func (rs *RecommendationService) GetWatchHistory(userID, limit int) ([]models.WatchHistory, error) {
	return rs.store.GetWatchHistory(userID, limit)
}
