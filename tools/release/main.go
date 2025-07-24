package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) && strings.Contains(string(exitError.Stderr), "No names found") {
			return "", nil
		}

		return "", fmt.Errorf("failed to run command '%s %s': %w\n%s", name, strings.Join(args, " "), err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

func executeStep(description string, command string, args ...string) {
	fmt.Println(description)

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	lastTag, err := runCmd("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting last tag: %v\n", err)
		os.Exit(1)
	}

	if lastTag == "" {
		fmt.Println("No tags found.")
	} else {
		fmt.Printf("Latest tag: %s\n", lastTag)
	}

	fmt.Print("Enter new tag: ")

	reader := bufio.NewReader(os.Stdin)

	newTag, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read from stdin: %v\n", err)
		os.Exit(1)
	}

	newTag = strings.TrimSpace(newTag)

	if newTag == "" {
		fmt.Println("No tag entered, aborting.")
		os.Exit(1)
	}

	if !strings.HasPrefix(newTag, "v") {
		newTag = "v" + newTag
	}

	executeStep(fmt.Sprintf("Tagging %s...", newTag), "git", "tag", newTag)

	majorVersion := ""

	parts := strings.Split(strings.TrimPrefix(newTag, "v"), ".")
	if len(parts) > 0 {
		majorVersion = "v" + parts[0]
	}

	if majorVersion != "" && majorVersion != newTag {
		executeStep(fmt.Sprintf("Updating major tag %s...", majorVersion), "git", "tag", "-f", majorVersion, newTag)
	}

	executeStep("Pushing tags...", "git", "push", "--tags", "--force")

	fmt.Printf("Successfully tagged and pushed %s.\n", newTag)
}
