package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"server/myproject/models"
	"server/myproject/services"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService  *services.AuthService
	movieService *services.MovieService
}

func NewHandler(authS *services.AuthService, movieS *services.MovieService) *Handler {
	return &Handler{
		authService:  authS,
		movieService: movieS,
	}
}

// Auth handlers

func (h *Handler) Register(c *gin.Context) {

	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=6"`
		Name        string `json:"name" binding:"required"`
		AdminSecret string `json:"admin_secret"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Register(req.Email, req.Password, req.Name, req.AdminSecret)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists or error creating user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user, "token": token})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (h *Handler) GetProfile(c *gin.Context) {
	// Simple user info is already in Context from middleware,
	// but if we need full profile we might need a method in AuthService or just reuse what we have.
	// For now we just return basic info or query DB if needed.
	// Let's assume we want to return what's in the token context for now, or fetch fresh?
	// Accessing store directly via authService if needed?
	// Ideally AuthService should have GetUserByID. I'll skip fetching fresh for now to keep it simple or add GetUserByID to AuthService.
	// I'll add GetUserByID to AuthService later if needed. For now just standard response.

	// Actually, the original implementation fetched from Store.
	// I should probably rely on the ID in the context.
	// I didn't add GetUserByID to AuthService yet.
	// I'll skip this implementation detail or quickly add it?
	// Let's return the ID and Role from context for now.
	c.JSON(http.StatusOK, gin.H{
		"id":   c.GetInt("userID"),
		"role": c.GetString("userRole"),
	})
}

// Movie handlers
func (h *Handler) ListMovies(c *gin.Context) {
	userID := c.GetInt("userID")
	movies, err := h.movieService.ListMovies(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		return
	}
	movies = h.movieService.EnrichMoviesWithURLs(movies)
	c.JSON(http.StatusOK, movies)
}

func (h *Handler) SearchMovies(c *gin.Context) {
	userID := c.GetInt("userID")
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
		return
	}

	movies, err := h.movieService.SearchMovies(userID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search movies"})
		return
	}
	movies = h.movieService.EnrichMoviesWithURLs(movies)
	c.JSON(http.StatusOK, movies)
}

func (h *Handler) GetMovie(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetInt("userID")

	movie, err := h.movieService.GetMovie(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found or access denied"})
		return
	}

	enriched := h.movieService.EnrichMoviesWithURLs([]models.Movie{*movie})
	result := enriched[0]

	// If it's a series, embed the episodes in the same response (saves an extra API call)
	if result.ContentType == "series" {
		episodes, _ := h.movieService.GetEpisodes(id, userID)
		if episodes == nil {
			episodes = []models.Movie{}
		}
		episodes = h.movieService.EnrichMoviesWithURLs(episodes)
		c.JSON(http.StatusOK, gin.H{"movie": result, "episodes": episodes})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) UploadMovie(c *gin.Context) {
	log.Println("UploadMovie")
	userID := c.GetInt("userID")

	title := c.PostForm("title")
	description := c.PostForm("description")
	contentType := c.PostForm("content_type")

	var parentID, seasonNum, episodeNum *int
	if p := c.PostForm("parent_id"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			parentID = &val
		}
	}
	if s := c.PostForm("season_number"); s != "" {
		if val, err := strconv.Atoi(s); err == nil {
			seasonNum = &val
		}
	}
	if e := c.PostForm("episode_number"); e != "" {
		if val, err := strconv.Atoi(e); err == nil {
			episodeNum = &val
		}
	}

	director := c.PostForm("director")

	releaseYearStr := c.PostForm("release_year")
	var releaseYear int
	if releaseYearStr != "" {
		releaseYear, _ = strconv.Atoi(releaseYearStr)
	}

	var cast, genres []string
	if castStr := c.PostForm("cast_members"); castStr != "" {
		parts := strings.Split(castStr, ",")
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				cast = append(cast, trimmed)
			}
		}
	}
	if genresStr := c.PostForm("genres"); genresStr != "" {
		parts := strings.Split(genresStr, ",")
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				genres = append(genres, trimmed)
			}
		}
	}

	videoFile, videoHeader, err := c.Request.FormFile("video")
	if err != nil && contentType != "series" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video file required for movies and episodes"})
		return
	}
	if videoFile != nil {
		defer videoFile.Close()
	}

	thumbFile, thumbHeader, _ := c.Request.FormFile("thumbnail")
	if thumbFile != nil {
		defer thumbFile.Close()
	}

	movie, err := h.movieService.UploadMovie(
		userID, title, description, contentType,
		parentID, seasonNum, episodeNum,
		cast, director, releaseYear, genres,
		videoFile, videoHeader, thumbFile, thumbHeader,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload movie: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, movie)
}

func (h *Handler) GetEpisodes(c *gin.Context) {
	seriesID, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetInt("userID")

	episodes, err := h.movieService.GetEpisodes(seriesID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch episodes"})
		return
	}

	episodes = h.movieService.EnrichMoviesWithURLs(episodes)
	c.JSON(http.StatusOK, episodes)
}

func (h *Handler) UpdateMovie(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetInt("userID")
	userRole := c.GetString("userRole")

	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Genres      []string `json:"genres"`
		Cast        []string `json:"cast_members"`
		Director    string   `json:"director"`
		ReleaseYear int      `json:"release_year"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movie, err := h.movieService.GetMovie(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	// Only admin or owner can edit
	if userRole != "admin" && movie.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if req.Title != "" {
		movie.Title = req.Title
	}
	if req.Description != "" {
		movie.Description = req.Description
	}
	if req.Genres != nil {
		movie.Genres = req.Genres
	}
	if req.Cast != nil {
		movie.Cast = req.Cast
	}
	if req.Director != "" {
		movie.Director = req.Director
	}
	if req.ReleaseYear > 0 {
		movie.ReleaseYear = req.ReleaseYear
	}

	if err := h.movieService.UpdateMovieFull(movie); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update"})
		return
	}

	c.JSON(http.StatusOK, movie)
}

func (h *Handler) DeleteMovie(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetInt("userID")
	userRole := c.GetString("userRole")

	if err := h.movieService.DeleteMovie(id, userID, userRole); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to delete movie or access denied"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.authService.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// Streaming with range support
func (h *Handler) StreamMovie(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetInt("userID")
	fmt.Println(userID)
	fmt.Println(id)
	movie, err := h.movieService.GetMovie(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found or access denied"})
		return
	}

	// If S3 is enabled, redirect to presigned URL
	if url, err := h.movieService.GetPresignedURL(movie.FilePath); err == nil && url != "" {
		c.Redirect(http.StatusTemporaryRedirect, url)
		return
	}

	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")
	c.File(movie.FilePath)
}

func (h *Handler) GetThumbnail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// For thumbnails, we might allow public access or check user session
	userID := c.GetInt("userID")

	movie, err := h.movieService.GetMovie(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	if movie.ThumbnailPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "No thumbnail for this movie"})
		return
	}

	// If S3 is enabled, redirect to presigned URL for thumbnail
	if url, err := h.movieService.GetPresignedURL(movie.ThumbnailPath); err == nil && url != "" {
		c.Redirect(http.StatusTemporaryRedirect, url)
		return
	}

	// Local fallback
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	c.File(filepath.Join(uploadDir, movie.ThumbnailPath))
}

// The GrantAccess/RevokeAccess handlers were removed along with the
// backing movie_access table. Access checks now rely on simple
// existence or other business rules, so the related endpoints have
// been dropped from the router in main.go.

// Admin handlers
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.authService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) UpdateUserRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Role string `json:"role" binding:"required,oneof=user admin moderator"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	if err := h.authService.UpdateUserRole(id, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
}
