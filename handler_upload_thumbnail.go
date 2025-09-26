package main

import (
	// "encoding/base64"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusBadRequest, "Unable to read file data", err)
	// 	return
	// }

	metaData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find video", err)
		return
	}
	if metaData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return
	}

	fileType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type header", err)
		return
	}
	fileExtension := ""
	switch fileType {
	case "image/jpeg":
		fileExtension = ".jpg"
	case "image/png":
		fileExtension = ".png"
	case "image/gif":
		fileExtension = ".gif"
	case "image/webp":
		fileExtension = ".webp"
	default:
		respondWithError(w, http.StatusUnsupportedMediaType, "Unsupported thumbnail type", nil)
		return
	}

	random := make([]byte, 32)
	rand.Read(random)
	randomName := base64.RawURLEncoding.EncodeToString(random)

	// fileName := videoIDString + fileExtension
	fileName := randomName + fileExtension
	thumbnailPath := filepath.Join(cfg.assetsRoot, fileName)

	thumbnailFile, err := os.Create(thumbnailPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	io.Copy(thumbnailFile, file)

	// thumbnail := thumbnail{
	// 	data:      data,
	// 	mediaType: mediaType,
	// }

	// videoThumbnails[videoID] = thumbnail
	// thumbnailURL := fmt.Sprintf("http://localhost:8091/api/thumbnails/%v", videoID)
	// metaData.ThumbnailURL = &thumbnailURL

	// thumbnailBase64 := base64.StdEncoding.EncodeToString(data)
	// dataURL := fmt.Sprintf("data:%s;base64,%s", mediaType, thumbnailBase64)
	// metaData.ThumbnailURL = &dataURL

	thumbnailURL := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, fileName)
	metaData.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(metaData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	respondWithJSON(w, http.StatusOK, metaData)
}
