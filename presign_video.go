package main

// import (
// 	"context"
// 	"fmt"
// 	"strings"
// 	"time"

// 	"github.com/aws/aws-sdk-go-v2/service/s3"
// 	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
// )

// func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
// 	presignedClient := s3.NewPresignClient(s3Client)

// 	getObjectInput := s3.GetObjectInput{
// 		Bucket: &bucket,
// 		Key:    &key,
// 	}

// 	presignedRequest, err := presignedClient.PresignGetObject(context.Background(), &getObjectInput, s3.WithPresignExpires(expireTime))
// 	if err != nil {
// 		return "", fmt.Errorf("error getting presigned request")
// 	}

// 	return presignedRequest.URL, nil
// }

// func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
// 	if video.VideoURL == nil || *video.VideoURL == "" {
// 		return video, nil
// 	}

// 	splitURL := strings.Split(*video.VideoURL, ",")
// 	if len(splitURL) != 2 {
// 		return video, nil
// 	}

// 	expiration := time.Minute * 5
// 	presignedURL, err := generatePresignedURL(cfg.s3Client, splitURL[0], splitURL[1], expiration)
// 	if err != nil {
// 		return video, err
// 	}

// 	video.VideoURL = &presignedURL

// 	return video, nil
// }
