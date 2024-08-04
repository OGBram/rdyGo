package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/sirupsen/logrus"
)

func main() {
	// Set up the logger
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Get input file from user
	fmt.Print("Enter the path to the input video file: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	inputFile := scanner.Text()

	// Generate the output file name
	outputFile := generateOutputFileName(inputFile)

	// Create a spinner for progress
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()

	// Build the FFmpeg command with the simplified progress bar
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-vf",
		"drawbox=y=ih-80:color=yellow:width=iw:height=80:t=fill",
		"-c:a", "copy",
		outputFile)

	// Print the command for debugging
	fmt.Printf("Running command: %s\n", strings.Join(cmd.Args, " "))

	// Capture FFmpeg's stderr to track progress
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logrus.Fatalf("Error creating stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		logrus.Fatalf("Error starting command: %v", err)
	}

	// Monitor FFmpeg progress
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "frame=") {
				// Update spinner based on progress info
				s.Suffix = " Processing: " + line
			}
		}
		if err := scanner.Err(); err != nil {
			logrus.Errorf("Error reading stderr: %v", err)
		}
	}()

	if err := cmd.Wait(); err != nil {
		s.Stop()
		logrus.Fatalf("Error executing command: %v", err)
	}

	s.Stop()
	fmt.Printf("Video processing completed successfully. Output saved to %s\n", outputFile)
}

// generateOutputFileName creates an output file name by adding an "output" suffix
func generateOutputFileName(inputFile string) string {
	ext := filepath.Ext(inputFile)
	name := strings.TrimSuffix(inputFile, ext)
	return fmt.Sprintf("%s_output%s", name, ext)
}
