package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type VideoInfo struct {
	ID    string
	Title string
	URL   string
}

func main() {
	log.Println("Starting...")
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if !checkDependencies() {
		log.Fatal("Required dependencies are missing. Please install yt-dlp.")
	}

	if len(config.Subscriptions) == 0 {
		log.Println("No subscriptions found in the configuration.")
		return
	}

	// Create archive directory
	if err := os.MkdirAll(ArchivesDir, 0755); err != nil {
		log.Printf("Error creating archive directory %s: %v", config.ArchivesDir, err)
		return
	}

	for _, sub := range config.Subscriptions {
		handleSubscription(sub, config)
	}
}
func checkDependencies() bool {
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return false
	}
	return true
}

func handleSubscription(sub Subscription, config *Config) {
	log.Printf("Handling subscription: %s", sub.Name)

	archiveFile := filepath.Join(config.ArchivesDir, fmt.Sprintf("%s.txt", sub.Name))

	// Ensure destination directory exists
	if err := os.MkdirAll(sub.Destination, 0755); err != nil {
		log.Printf("Error creating destination directory %s: %v", sub.Destination, err)
		return
	}

	videos, err := getVideoList(sub)
	if err != nil {
		log.Printf("Error fetching video list for %s: %v", sub.Name, err)
		return
	}

	log.Printf("Found %d videos for subscription: %s", len(videos), sub.Name)

	// Apply filters and check against archive
	filteredVideos, filteredCount, skippedCount := filterVideos(videos, sub, archiveFile)

	if len(filteredVideos) == 0 {
		log.Printf("No videos to download for subscription: %s (Filtered: %d, Skipped: %d)", sub.Name, filteredCount, skippedCount)
		return
	}

	// check number of episodes already in destination
	files, err := os.ReadDir(sub.Destination)
	if err != nil {
		log.Printf("Error reading destination directory %s: %v", sub.Destination, err)
		return
	}
	episodesCount := 0
	for _, file := range files {
		matched, err := regexp.MatchString(`.*E\d+.*\.mp4$`, file.Name())
		if err != nil {
			log.Printf("Error matching file name %s: %v", file.Name(), err)
			continue
		}
		if matched {
			episodesCount++
		}
	}

	downloadCount := 0
	for _, video := range filteredVideos {
		success := downloadVideo(video, episodesCount+downloadCount+1, sub, archiveFile)
		if success {
			downloadCount++
			log.Printf("Downloaded video: %s (Title: %s)", video.ID, video.Title)
		} else {
			log.Printf("Failed to download video: %s (Title: %s)", video.ID, video.Title)
		}
	}

	log.Printf("Subscription %s completed: Downloaded %d videos, Filtered %d, Skipped %d", sub.Name, downloadCount, filteredCount, skippedCount)
	return
}

func getVideoList(sub Subscription) ([]VideoInfo, error) {
	// Use yt-dlp to get only specific fields in a custom format
	cmd := exec.Command("yt-dlp",
		"--print", "%(id)s:|:%(title)s:|:%(webpage_url)s",
		"--playlist-end", strconv.Itoa(sub.MaxVideos),
		"--ignore-errors",
		"--no-warnings",
		"--quiet",
		sub.URL,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video list: %v", err)
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("no output from yt-dlp")
	}

	var videos []VideoInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split by the pipe delimiter
		parts := strings.Split(line, ":|:")
		if len(parts) != 3 {
			fmt.Printf("Warning: line %d has %d parts instead of 3: %s\n", i, len(parts), line)
			continue
		}

		video := VideoInfo{
			ID:    parts[0],
			Title: parts[1],
			URL:   parts[2],
		}

		videos = append(videos, video)
	}

	return videos, nil
}

func filterVideos(videos []VideoInfo, sub Subscription, archiveFile string) ([]VideoInfo, int, int) {
	var filteredVideos []VideoInfo
	filteredCount := 0
	skippedCount := 0

	// Load existing archive entries
	downloadedVideos := loadArchiveEntries(archiveFile)

	for _, video := range videos {
		// Check if already downloaded
		if _, exists := downloadedVideos[video.ID]; exists {
			skippedCount++
			log.Printf("Skipping already downloaded video: %s (Title: %s)", video.ID, video.Title)
			continue
		}

		// Apply include/exclude patterns
		if shouldDownloadVideo(video, sub) {
			filteredVideos = append(filteredVideos, video)
			log.Printf("Video passed filter: %s (Title: %s)", video.ID, video.Title)
		} else {
			filteredCount++
		}
	}

	return filteredVideos, filteredCount, skippedCount
}

func shouldDownloadVideo(video VideoInfo, sub Subscription) bool {
	// If no pattern specified, download all
	if sub.Filter == "" {
		return true
	}

	// Check include patterns
	titleMatch, titleErr := regexp.MatchString(sub.Filter, video.Title)
	if titleErr != nil {
		log.Printf("Error matching title pattern: %v", titleErr)
		return false
	}
	return titleMatch
}

func loadArchiveEntries(archiveFile string) map[string]bool {
	entries := make(map[string]bool)

	file, err := os.Open(archiveFile)
	if err != nil {
		return entries // Return empty map if file doesn't exist
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// Archive format is usually "youtube VIDEOID"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				entries[parts[1]] = true
			}
		}
	}

	return entries
}

func downloadVideo(video VideoInfo, episodeNumber int, sub Subscription, archiveFile string) bool {
	filename := fmt.Sprintf("%s E%03d", sub.Name, episodeNumber)+" [%(id)s].%(ext)s"
	if sub.FilenameTemplate != "" {
		filename = sub.FilenameTemplate
	}
	cmd := exec.Command("yt-dlp",
		"--download-archive", archiveFile,
		"--output", filepath.Join(DownloadsDir, sub.Destination, filename),
		"--format", "best[height<=1080]",
		"--no-overwrites",
		"--continue",
		"--ignore-errors",
		"--embed-thumbnail",
		"--add-metadata",
		video.URL,
	)

	output, err := cmd.CombinedOutput()

	// Log any output
	if len(output) > 0 {
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				log.Println(line)
			}
		}
	}

	return err == nil
}
