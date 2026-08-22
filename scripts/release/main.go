package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	args := []string{"tag", "--list", "v*.*.*", "--sort=-v:refname"}
	cmd := exec.Command("git", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, output)
	}

	lastTag := strings.TrimSpace(string(output))

	if first, _, found := strings.Cut(lastTag, "\n"); found {
		lastTag = first
	}

	if lastTag == "" {
		fmt.Println("No semantic version tags found (e.g., v1.2.3).")
	} else {
		fmt.Printf("Latest tag: %s\n", lastTag)
	}

	fmt.Print("Enter new tag: ")

	newTag, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("read new tag: %w", err)
	}

	newTag = strings.TrimSpace(newTag)
	if newTag == "" {
		return errors.New("no tag entered")
	}

	if !strings.HasPrefix(newTag, "v") {
		newTag = "v" + newTag
	}

	fmt.Printf("Tagging %s...\n", newTag)

	if err := executeGit("tag", newTag); err != nil {
		return err
	}

	major, _, _ := strings.Cut(strings.TrimPrefix(newTag, "v"), ".")
	majorTag := "v" + major
	updateMajorTag := major != "" && majorTag != newTag

	if updateMajorTag {
		fmt.Printf("Updating major tag %s...\n", majorTag)

		if err := executeGit("tag", "-f", majorTag, newTag); err != nil {
			return err
		}
	}

	fmt.Printf("Pushing %s...\n", newTag)

	if err := executeGit("push", "origin", newTag); err != nil {
		return err
	}

	if updateMajorTag {
		fmt.Printf("Pushing %s...", majorTag)

		if err := executeGit("push", "--force", "origin", majorTag); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully tagged and pushed %s.\n", newTag)

	return nil
}

func executeGit(args ...string) error {
	cmd := exec.Command("git", args...) //nolint:gosec // Arguments are passed directly to Git without a shell.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return nil
}
