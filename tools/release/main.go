package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunCMD(name string, args ...string) (string, error) {
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

func ExecuteStep(description string, command string, args ...string) {
	fmt.Println(description)

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func main() {
	lastTag, err := RunCMD("git", "tag", "--list", "v*.*.*", "--sort=-v:refname")
	if err != nil {
		panic(err)
	}

	if i := strings.Index(lastTag, "\n"); i != -1 {
		lastTag = lastTag[:i]
	}

	if lastTag == "" {
		fmt.Println("No semantic version tags found (e.g., v1.2.3).")
	} else {
		fmt.Printf("Latest tag: %s\n", lastTag)
	}

	fmt.Print("Enter new tag: ")

	newTag, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}

	newTag = strings.TrimSpace(newTag)

	if newTag == "" {
		panic("No tag entered, aborting.")
	}

	if !strings.HasPrefix(newTag, "v") {
		newTag = "v" + newTag
	}

	ExecuteStep(fmt.Sprintf("Tagging %s...", newTag), "git", "tag", newTag)

	majorVersion := ""

	if parts := strings.Split(strings.TrimPrefix(newTag, "v"), "."); len(parts) > 0 {
		majorVersion = "v" + parts[0]
	}

	if majorVersion != "" && majorVersion != newTag {
		ExecuteStep(fmt.Sprintf("Updating major tag %s...", majorVersion), "git", "tag", "-f", majorVersion, newTag)
	}

	ExecuteStep(fmt.Sprintf("Pushing %s...", newTag), "git", "push", "origin", newTag)

	if majorVersion != "" && majorVersion != newTag {
		ExecuteStep(fmt.Sprintf("Pushing %s...", majorVersion), "git", "push", "--force", "origin", majorVersion)
	}

	fmt.Printf("Successfully tagged and pushed %s.\n", newTag)
}
