package main

import (
	"net/http"
	"strconv"

	"server/myproject/models"
	"server/myproject/services"

	"github.com/gin-gonic/gin"
)

type RecommendationHandler struct {
	service      *services.RecommendationService
	movieService *services.MovieService
}

func NewRecommendationHandler(service *services.RecommendationService, movieService *services.MovieService) *RecommendationHandler {
	return &RecommendationHandler{service: service, movieService: movieService}
}

// GET /api/movies/:id/similar?limit=5
func (h *RecommendationHandler) GetSimilarMovies(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	userID := c.GetInt("userID")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	// Check access
	if _, err := h.movieService.GetMovie(movieID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied or movie not found"})
		return
	}

	movies, err := h.service.GetSimilarMovies(movieID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get similar movies",
			"details": err.Error(),
		})
		return
	}

	// Filter to only movies user can access
	var accessible []models.Movie
	for _, m := range movies {
		// Efficiently check access?
		// Calling GetMovie in loop might be slow if it hits DB each time.
		// But for "similar" list (limit 5), it's fine.
		if _, err := h.movieService.GetMovie(m.ID, userID); err == nil {
			accessible = append(accessible, m)
		}
	}

	// Enrich accessible movies
	accessible = h.movieService.EnrichMoviesWithURLs(accessible)

	c.JSON(http.StatusOK, gin.H{
		"movie_id": movieID,
		"similar":  accessible,
		"count":    len(accessible),
	})
}

// POST /api/watch-history
func (h *RecommendationHandler) UpdateWatchProgress(c *gin.Context) {
	userID := c.GetInt("userID")

	var req struct {
		MovieID  int `json:"movie_id" binding:"required"`
		Progress int `json:"progress" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check access
	if _, err := h.movieService.GetMovie(req.MovieID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.service.UpdateWatchProgress(userID, req.MovieID, req.Progress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Progress updated"})
}

// GET /api/watch-history?limit=20
func (h *RecommendationHandler) GetWatchHistory(c *gin.Context) {
	userID := c.GetInt("userID")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	history, err := h.service.GetWatchHistory(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get history"})
		return
	}

	// Enrich with movie details
	type HistoryItem struct {
		models.WatchHistory
		Movie *models.Movie `json:"movie"`
	}

	var items []HistoryItem
	for _, wh := range history {
		item := HistoryItem{WatchHistory: wh}
		// Fetch movie details
		if movie, err := h.movieService.GetMovie(wh.MovieID, userID); err == nil {
			item.Movie = movie
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"history": items,
		"count":   len(items),
	})
}

// GET /api/browse
func (h *RecommendationHandler) GetBrowseData(c *gin.Context) {
	userID := c.GetInt("userID")

	// 1. All Movies
	allMovies, err := h.movieService.ListMovies(userID)
	if err != nil {
		allMovies = []models.Movie{} // Fallback to empty
	}

	// 2. Recommendations
	recommendations, err := h.service.GetRecommendations(userID, 10)
	if err != nil {
		recommendations = []models.Movie{}
	}

	// 3. Watch History
	history, err := h.service.GetWatchHistory(userID, 10)
	if err != nil {
		history = []models.WatchHistory{}
	}

	// Enrich history with movie details
	type HistoryItem struct {
		models.WatchHistory
		Movie *models.Movie `json:"movie"`
	}
	var enrichedHistory []HistoryItem
	for _, wh := range history {
		item := HistoryItem{WatchHistory: wh}
		if movie, err := h.movieService.GetMovie(wh.MovieID, userID); err == nil {
			// Enrich the single movie in the history item
			enriched := h.movieService.EnrichMoviesWithURLs([]models.Movie{*movie})
			item.Movie = &enriched[0]
		}
		enrichedHistory = append(enrichedHistory, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"all_movies":      h.movieService.EnrichMoviesWithURLs(allMovies),
		"recommendations": h.movieService.EnrichMoviesWithURLs(recommendations),
		"watch_history":   enrichedHistory,
	})
}
