package image

import (
	"bufio"
	"os"
	"strings"
)

type ImageList struct {
	Images map[string][]string
}

func LoadFromFile(filepath string) (*ImageList, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	images := make(map[string][]string)
	var currentSection string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			if _, ok := images[currentSection]; !ok {
				images[currentSection] = []string{}
			}
			continue
		}

		if currentSection == "" {
			currentSection = "default"
			if _, ok := images[currentSection]; !ok {
				images[currentSection] = []string{}
			}
		}

		if idx := strings.Index(line, "@"); idx != -1 {
			line = line[:idx]
		}

		if line != "" {
			images[currentSection] = append(images[currentSection], line)
		}
	}

	return &ImageList{Images: images}, scanner.Err()
}

func (il *ImageList) GetImages(provider string) []string {
	if images, ok := il.Images[provider]; ok {
		return images
	}
	return []string{}
}

func (il *ImageList) Count(provider string) int {
	return len(il.GetImages(provider))
}
