package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/soli0222/misskey-summarizer/internal/discord"
	"github.com/soli0222/misskey-summarizer/internal/misskey"
	"github.com/soli0222/misskey-summarizer/internal/openai"
	"github.com/spf13/cobra"
)

var (
	// Flags
	dateStr       string
	yesterday     bool
	last24h       bool
	outputFormat  string
	postToDiscord bool

	rootCmd = &cobra.Command{
		Use:   "misskey-summarizer",
		Short: "Summarize your daily Misskey notes using AI",
		Long: `A CLI tool that fetches your Misskey notes for a specific day,
generates a summary using OpenAI's API, and optionally posts it to Discord.

Example:
  misskey-summarizer --yesterday --discord`,
		RunE: runSummarizer,
	}
)

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&dateStr, "date", "d", "", "Target date in YYYY-MM-DD format")
	rootCmd.Flags().BoolVarP(&yesterday, "yesterday", "y", false, "Use yesterday's date")
	rootCmd.Flags().BoolVarP(&last24h, "last24h", "l", false, "Use the last 24 hours from now")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "summary", "Output format: summary or json")
	rootCmd.Flags().BoolVar(&postToDiscord, "discord", false, "Post summary to Discord webhook")
}

// Config holds the application configuration
type Config struct {
	MisskeyToken       string
	MisskeyInstanceURL string
	OpenAIAPIKey       string
	OpenAIModel        string
	DiscordWebhookURL  string
}

func runSummarizer(cmd *cobra.Command, args []string) error {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine target time range
	startTime, endTime, err := determineTargetTimeRange(dateStr, yesterday, last24h)
	if err != nil {
		return fmt.Errorf("failed to determine target time range: %w", err)
	}

	if last24h {
		log.Printf("Fetching notes from %s to %s...", startTime.Format("2006-01-02 15:04"), endTime.Format("2006-01-02 15:04"))
	} else {
		log.Printf("Fetching notes for %s...", startTime.Format("2006-01-02"))
	}

	// Create Misskey client and get user info
	misskeyClient := misskey.NewClient(config.MisskeyInstanceURL, config.MisskeyToken)

	me, err := misskeyClient.GetMe()
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}
	log.Printf("Authenticated as: @%s", me.Username)

	// Fetch notes for the target time range
	notes, err := misskeyClient.GetNotesForTimeRange(me.ID, startTime, endTime, true) // includeRenotes = true
	if err != nil {
		return fmt.Errorf("failed to fetch notes: %w", err)
	}

	log.Printf("Fetched %d notes", len(notes))

	// Extract text content from notes
	var noteTexts []string
	for _, note := range notes {
		text := note.GetDisplayText()
		if text != "" {
			noteTexts = append(noteTexts, text)
		}
	}

	// Generate summary using OpenAI
	openaiClient := openai.NewClient(config.OpenAIAPIKey)
	if config.OpenAIModel != "" {
		openaiClient.WithModel(config.OpenAIModel)
	}

	var dateFormatted string
	if last24h {
		dateFormatted = fmt.Sprintf("%s 〜 %s", startTime.Format("1月2日 15:04"), endTime.Format("1月2日 15:04"))
	} else {
		dateFormatted = startTime.Format("2006年1月2日")
	}
	summary, err := openaiClient.SummarizeNotes(noteTexts, dateFormatted)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	// Output results
	switch outputFormat {
	case "none":
		log.Println("Output format set to 'none', skipping output")
	case "json":
		output := map[string]interface{}{
			"start_time": startTime.Format(time.RFC3339),
			"end_time":   endTime.Format(time.RFC3339),
			"note_count": len(notes),
			"summary":    summary,
			"notes":      noteTexts,
		}
		jsonOutput, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonOutput))
	default:
		fmt.Printf("\n📅 %s のサマリー\n", dateFormatted)
		fmt.Printf("📊 ノート数: %d\n\n", len(notes))
		fmt.Println(summary)
	}

	// Post to Discord if requested
	if postToDiscord {
		if config.DiscordWebhookURL == "" {
			log.Println("Warning: DISCORD_WEBHOOK_URL is not set, skipping Discord post")
		} else {
			discordClient := discord.NewClient(config.DiscordWebhookURL)
			if err := discordClient.PostSummary(dateFormatted, len(notes), summary); err != nil {
				log.Printf("Failed to post to Discord: %v", err)
			} else {
				log.Println("Posted summary to Discord")
			}
		}
	}

	return nil
}

// loadConfig loads configuration from environment variables
func loadConfig() (*Config, error) {
	config := &Config{
		MisskeyToken:       os.Getenv("MISSKEY_TOKEN"),
		MisskeyInstanceURL: os.Getenv("MISSKEY_INSTANCE_URL"),
		OpenAIAPIKey:       os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:        os.Getenv("OPENAI_MODEL"),
		DiscordWebhookURL:  os.Getenv("DISCORD_WEBHOOK_URL"),
	}

	if config.MisskeyToken == "" {
		return nil, fmt.Errorf("MISSKEY_TOKEN environment variable is required")
	}
	if config.MisskeyInstanceURL == "" {
		return nil, fmt.Errorf("MISSKEY_INSTANCE_URL environment variable is required")
	}
	if config.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	return config, nil
}

// determineTargetTimeRange determines the target time range based on flags
func determineTargetTimeRange(dateStr string, useYesterday bool, useLast24h bool) (time.Time, time.Time, error) {
	if useLast24h {
		now := time.Now()
		return now.Add(-24 * time.Hour), now, nil
	}

	var targetDate time.Time
	if dateStr != "" {
		parsed, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
		}
		targetDate = parsed
	} else if useYesterday {
		targetDate = time.Now().AddDate(0, 0, -1)
	} else {
		// Default to today
		targetDate = time.Now()
	}

	// Get start and end of day
	loc := targetDate.Location()
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	return startOfDay, endOfDay, nil
}
