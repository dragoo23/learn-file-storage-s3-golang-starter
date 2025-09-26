package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

type aspectRatio struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("error running command: ", err)
		fmt.Println("stderr: ", stderr.String())
		return "", fmt.Errorf("error getting video ascept ratio: %w", err)
	}

	var data aspectRatio
	decoder := json.NewDecoder(&out)
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("error decoding json")
		return "", fmt.Errorf("error decoding json data: %w", err)
	}

	return decideAspectRatio(data)
}

func decideAspectRatio(data aspectRatio) (string, error) {
	dimensions := data.Streams[0]
	if dimensions.Height == 0 || dimensions.Width == 0 {
		fmt.Println("0 values in dimensions")
		return "", fmt.Errorf("missing video dimensions")
	}

	ratio := float64(dimensions.Width) / float64(dimensions.Height)
	const tolerance = 0.01
	if math.Abs(ratio-(16.0/9.0)) <= tolerance {
		return "16:9", nil
	} else if math.Abs(ratio-(9.0/16.0)) <= tolerance {
		return "9:16", nil
	} else {
		return "other", nil
	}
}
