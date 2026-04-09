package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/pengshtime/docker-image-sync/internal/config"
	"github.com/pengshtime/docker-image-sync/internal/image"
	"github.com/pengshtime/docker-image-sync/internal/provider"
)

func main() {
	cfg := config.Load()

	images, err := image.LoadFromFile(cfg.ImageList)
	if err != nil {
		log.Fatalf("Failed to load image list: %v", err)
	}

	providerImages := images.GetImages(cfg.Provider)
	totalImages := len(providerImages)

	if totalImages == 0 {
		fmt.Printf("No images found for provider: %s\n", cfg.Provider)
		os.Exit(0)
	}

	fmt.Printf("Loaded %d images for provider: %s\n", totalImages, cfg.Provider)

	factory := provider.NewProviderFactory()
	p, err := factory.Create(cfg.Provider, cfg.Registry, cfg.Namespace, cfg.Username, cfg.Password)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Printf("Using provider: %s (%s)\n", p.Name(), p.RegistryDomain())

	ctx := context.Background()

	var wg sync.WaitGroup
	imageChan := make(chan string, totalImages)
	resultChan := make(chan *provider.SyncResult, totalImages)

	for _, img := range providerImages {
		imageChan <- img
	}
	close(imageChan)

	concurrency := 3
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for img := range imageChan {
				result, err := p.SyncImage(ctx, img)
				if err != nil {
					result = &provider.SyncResult{
						SourceImage:  img,
						Success:      false,
						ErrorMessage: err.Error(),
					}
				}
				resultChan <- result
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var successCount, failureCount, skippedCount int

	for result := range resultChan {
		if result.Success {
			if result.ErrorMessage == "already exists" {
				skippedCount++
				fmt.Printf("[SKIP] %s -> %s (already exists)\n", result.SourceImage, result.TargetImage)
			} else {
				successCount++
				fmt.Printf("[SUCCESS] %s -> %s\n", result.SourceImage, result.TargetImage)
			}
		} else {
			failureCount++
			fmt.Printf("[FAIL] %s -> %s: %s\n", result.SourceImage, result.TargetImage, result.ErrorMessage)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("Sync Summary")
	fmt.Println("========================================")
	fmt.Printf("Provider: %s\n", cfg.Provider)
	fmt.Printf("Total: %d\n", totalImages)
	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Skipped: %d\n", skippedCount)
	fmt.Printf("Failed: %d\n", failureCount)
	fmt.Println("========================================")

	if failureCount > 0 {
		os.Exit(1)
	}
}
