package main

import (
	"bulk-mail/internal/app"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Check if required files exist
	if !filesExist() {
		fmt.Println("Required files not found (config.yaml, data.txt, mail.html)")
		fmt.Print("Do you want to create sample files? (y/n): ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			if err := app.CreateSampleConfig(); err != nil {
				fmt.Printf("Error creating config: %v\n", err)
				os.Exit(1)
			}
			if err := app.CreateSampleTemplate("mail.html"); err != nil {
				fmt.Printf("Error creating template: %v\n", err)
				os.Exit(1)
			}
			if err := app.CreateSampleData("data.txt"); err != nil {
				fmt.Printf("Error creating data: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("\nSample files created successfully!")
			fmt.Println("Please edit config.yaml with your SMTP settings and restart the application.")
		}
		return
	}

	application := &app.App{}
	if err := application.Init(); err != nil {
		fmt.Printf("Init error: %v\n", err)
		os.Exit(1)
	}
	defer application.Watcher.Close()

	if err := app.RunTUI(application); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func filesExist() bool {
	files := []string{"config.yaml", "data.txt", "mail.html"}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return false
		}
	}
	return true
}
