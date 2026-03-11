package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"server/myproject/models"
	"server/myproject/store"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type MovieService struct {
	store    *store.Store
	s3Client *s3.Client
	bucket   string
}

func NewMovieService(s *store.Store) *MovieService {
	return &MovieService{
		store:  s,
		bucket: os.Getenv("AWS_S3_BUCKET"),
	}
}

func (s *MovieService) SetS3Client(client *s3.Client) {
	s.s3Client = client
}

func (s *MovieService) ListMovies(userID int) ([]models.Movie, error) {
	movies, err := s.store.ListMoviesForUser(userID)
	if movies == nil {
		return []models.Movie{}, nil
	}
	return movies, err
}

func (s *MovieService) SearchMovies(userID int, query string) ([]models.Movie, error) {
	return s.store.SearchMovies(userID, query)
}

func (s *MovieService) GetMovie(id, userID int) (*models.Movie, error) {
	movie, err := s.store.GetMovie(id)
	if err != nil {
		return nil, err
	}

	// Resolve absolute path only for Local Storage
	if s.s3Client == nil || s.bucket == "" {
		uploadDir := os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			uploadDir = "./uploads"
		}
		movie.FilePath = filepath.Join(uploadDir, movie.FilePath)
	}

	return movie, nil
}

func (s *MovieService) UploadMovie(
	userID int,
	title, description, contentType string,
	parentID, seasonNum, episodeNum *int,
	cast []string, director string, releaseYear int, genres []string,
	videoFile multipart.File, videoHeader *multipart.FileHeader,
	thumbFile multipart.File, thumbHeader *multipart.FileHeader,
) (*models.Movie, error) {
	var videoPathDB, thumbPathDB string
	var filename string

	// 1. Video Logic (Optional for Series)
	if videoFile != nil && videoHeader != nil {
		ext := filepath.Ext(videoHeader.Filename)
		filename = fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)

		if s.s3Client != nil && s.bucket != "" {
			if err := s.uploadToS3(filename, videoFile); err != nil {
				return nil, fmt.Errorf("failed to upload video to S3: %v", err)
			}
			videoPathDB = filename
		} else {
			uploadDir := os.Getenv("UPLOAD_DIR")
			if uploadDir == "" {
				uploadDir = "./uploads"
			}
			os.MkdirAll(uploadDir, 0755)
			absolutePath := filepath.Join(uploadDir, filename)
			out, err := os.Create(absolutePath)
			if err != nil {
				return nil, err
			}
			defer out.Close()
			if _, err := io.Copy(out, videoFile); err != nil {
				return nil, err
			}
			videoPathDB = filename
		}
	}

	// 2. Thumbnail Logic (Optional)
	if thumbFile != nil && thumbHeader != nil {
		thumbExt := filepath.Ext(thumbHeader.Filename)
		thumbFilename := fmt.Sprintf("thumb_%s%s", uuid.New().String(), thumbExt)

		if s.s3Client != nil && s.bucket != "" {
			if err := s.uploadToS3(thumbFilename, thumbFile); err != nil {
				log.Printf("Failed to upload thumbnail to S3: %v", err)
			} else {
				thumbPathDB = thumbFilename
			}
		} else {
			uploadDir := os.Getenv("UPLOAD_DIR")
			if uploadDir == "" {
				uploadDir = "./uploads"
			}
			os.MkdirAll(uploadDir, 0755)
			absoluteThumbPath := filepath.Join(uploadDir, thumbFilename)
			thumbOut, err := os.Create(absoluteThumbPath)
			if err == nil {
				defer thumbOut.Close()
				io.Copy(thumbOut, thumbFile)
				thumbPathDB = thumbFilename
			}
		}
	}

	if contentType == "" {
		contentType = "movie"
	}

	movie := &models.Movie{
		Title:         title,
		Description:   description,
		FilePath:      videoPathDB,
		ThumbnailPath: thumbPathDB,
		OwnerID:       userID,
		ContentType:   contentType,
		ParentID:      parentID,
		SeasonNum:     seasonNum,
		EpisodeNum:    episodeNum,
		Cast:          cast,
		Director:      director,
		ReleaseYear:   releaseYear,
		Genres:        genres,
	}

	if err := s.store.CreateMovie(movie); err != nil {
		return nil, err
	}

	return movie, nil
}

func (s *MovieService) GetEpisodes(seriesID, userID int) ([]models.Movie, error) {
	// Simple access check for series exists (optional but good)
	_, err := s.store.GetMovie(seriesID)
	if err != nil {
		return nil, err
	}

	episodes, err := s.store.GetEpisodesForSeries(seriesID)
	if err != nil {
		return nil, err
	}
	return episodes, nil
}

func (s *MovieService) GetPresignedURL(key string) (string, error) {
	if s.s3Client == nil {
		return "", fmt.Errorf("S3 client not initialized")
	}

	presignClient := s3.NewPresignClient(s.s3Client)
	request, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 1 * time.Hour
	})

	if err != nil {
		return "", err
	}

	return request.URL, nil
}

// EnrichMoviesWithURLs attaches S3 presigned URLs for video and thumbnail to a slice of movies
func (s *MovieService) EnrichMoviesWithURLs(movies []models.Movie) []models.Movie {
	for i := range movies {
		// Generate Thumbnail URL
		if movies[i].ThumbnailPath != "" {
			if s.s3Client != nil {
				url, err := s.GetPresignedURL(movies[i].ThumbnailPath)
				if err == nil {
					movies[i].ThumbnailURL = url
				}
			} else {
				// Fallback to local
				movies[i].ThumbnailURL = "/uploads/" + movies[i].ThumbnailPath
			}
		}

		// Generate Video URL
		if movies[i].FilePath != "" {
			if s.s3Client != nil {
				url, err := s.GetPresignedURL(movies[i].FilePath)
				if err == nil {
					movies[i].VideoURL = url
				}
			} else {
				movies[i].VideoURL = "/api/stream/" + fmt.Sprintf("%d", movies[i].ID)
			}
		}
	}
	return movies
}

func (s *MovieService) uploadToS3(key string, file io.Reader) error {
	_, err := s.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   file,
		// ACL: "public-read", // Optional: depends on bucket policy
	})
	return err
}

func (s *MovieService) UpdateMovie(id, userID int, title, description string) (*models.Movie, error) {
	movie, err := s.store.GetMovie(id)
	if err != nil {
		return nil, err
	}

	movie.Title = title
	movie.Description = description

	if err := s.store.UpdateMovie(movie); err != nil {
		return nil, err
	}

	return movie, nil
}

// UpdateMovieFull saves all metadata fields (for admin edit)
func (s *MovieService) UpdateMovieFull(m *models.Movie) error {
	return s.store.UpdateMovieFull(m)
}

func (s *MovieService) DeleteMovie(id, userID int, userRole string) error {
	movie, err := s.store.GetMovie(id)
	if err != nil {
		return err
	}

	// Admin can delete anything, otherwise only the owner can
	if userRole != "admin" {
		return store.ErrAccessDenied
	}

	os.Remove(movie.FilePath)
	if movie.ThumbnailPath != "" {
		os.Remove(movie.ThumbnailPath)
	}

	return s.store.DeleteMovie(id)
}

// GrantAccess and RevokeAccess operations have been removed; access
// is no longer recorded in a dedicated table.  The frontend routes
// and handlers related to access control should also be cleaned up.
