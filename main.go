package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"main/utils"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Flags for server port
	port := flag.Int("port", 80, "Port to run the server on")
	flag.Parse()

	// Initiate App Config
	if err := startConfig(); err != nil {
		log.Fatalf("Failed to start configuration: %v", err)
	}

	// Register the HTTP handlers
	http.HandleFunc("/", ok)
	http.HandleFunc("/config", configResource)

	// Start the HTTP server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on port %v", *port)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func ok(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK\n")
}

var client *appconfigdata.Client
var token *string
var configStr string

func startConfig() error {
	// Start session and retrieve initial configuration token
	initialConfigurationToken, err := startSession()
	if err != nil {
		return err
	}
	token = initialConfigurationToken

	// Start a goroutine to get the latest configuration continuously
	go getLatestConfigInfiniteLoop()
	return nil
}

func startSession() (*string, error) {
	params := utils.GetParameters()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	log.Println("AWS Configuration loaded successfully.")

	minInterval := int32(15)

	client = appconfigdata.NewFromConfig(cfg)
	input := appconfigdata.StartConfigurationSessionInput{
		ApplicationIdentifier:                &params.AppId,
		ConfigurationProfileIdentifier:       &params.ConfigProfileId,
		EnvironmentIdentifier:                &params.EnvId,
		RequiredMinimumPollIntervalInSeconds: &minInterval,
	}

	output, err := client.StartConfigurationSession(context.TODO(), &input)
	if err != nil {
		return nil, fmt.Errorf("failed to start configuration session: %w", err)
	}

	log.Printf("Started configuration session with token: %s", *output.InitialConfigurationToken)

	return output.InitialConfigurationToken, nil
}

func getLatestConfigInfiniteLoop() {
	for {
		fmt.Println("Retrieving latest config...")
		input := appconfigdata.GetLatestConfigurationInput{
			ConfigurationToken: token,
		}

		output, err := client.GetLatestConfiguration(context.TODO(), &input)
		if err != nil {
			log.Printf("Error retrieving latest configuration: %v", err)
			time.Sleep(10 * time.Second) // Retry after a delay
			continue
		}

		token = output.NextPollConfigurationToken
		latest := output.Configuration
		if len(latest) != 0 {
			configStr = string(latest)
			log.Printf("New latest configuration retrieved: %s", configStr)
		} else {
			log.Println("Nothing changed, already using the latest configuration!")
		}

		interval := output.NextPollIntervalInSeconds
		duration := time.Second * time.Duration(interval)
		log.Printf("Sleeping for %s", duration)
		time.Sleep(duration)
	}
}

func configResource(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Latest Configuration: ", configStr, "\n")
}