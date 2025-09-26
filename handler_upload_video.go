package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)

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

	metaData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find video", err)
		return
	}
	if metaData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	fileType := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(fileType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type header", err)
		return
	}

	fileExtension := ""
	switch mediaType {
	case "video/mp4":
		fileExtension = ".mp4"
	default:
		respondWithError(w, http.StatusUnsupportedMediaType, "Unsupported video type", nil)
		return
	}

	tempFile, err := os.CreateTemp("", "tubely-upload-*.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	ratio, err := getVideoAspectRatio(tempFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	prefix := ""
	switch ratio {
	case "16:9":
		prefix = "landscape"
	case "9:16":
		prefix = "portrait"
	default:
		prefix = "other"
	}

	outputPath, err := processVideoForFastStart(tempFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	processedFile, err := os.Open(outputPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}
	defer os.Remove(processedFile.Name())
	defer processedFile.Close()

	random := make([]byte, 32)
	rand.Read(random)
	randomName := base64.RawURLEncoding.EncodeToString(random)
	key := prefix + "/" + randomName + fileExtension

	putObjectInput := s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &key,
		Body:        processedFile,
		ContentType: &mediaType,
	}

	_, err = cfg.s3Client.PutObject(r.Context(), &putObjectInput)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upload to S3", err)
		return
	}

	videoURL := fmt.Sprintf("https://%s/%s", cfg.s3CfDistribution, key)
	metaData.VideoURL = &videoURL

	err = cfg.db.UpdateVideo(metaData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	// video, err := cfg.dbVideoToSignedVideo(metaData)
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
	// 	return
	// }

	respondWithJSON(w, http.StatusOK, metaData)
}
