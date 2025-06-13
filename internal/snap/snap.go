package snap

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"uninn/pkg"
)

type SnapDetector struct{}

func New() *SnapDetector {
	return &SnapDetector{}
}

func (s *SnapDetector) IsAvailable() bool {
	_, err := exec.LookPath("snap")
	return err == nil
}

func (s *SnapDetector) ListPackages() ([]pkg.Package, error) {
	cmd := exec.Command("snap", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list snap packages: %w", err)
	}

	var packages []pkg.Package
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	// Skip header line
	if scanner.Scan() {
		scanner.Text()
	}
	
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 4 {
			// Get more info about the snap
			infoCmd := exec.Command("snap", "info", fields[0])
			infoOutput, _ := infoCmd.Output()
			
			var description, size string
			infoScanner := bufio.NewScanner(strings.NewReader(string(infoOutput)))
			for infoScanner.Scan() {
				line := infoScanner.Text()
				if strings.HasPrefix(line, "summary:") {
					description = strings.TrimSpace(strings.TrimPrefix(line, "summary:"))
				} else if strings.Contains(line, "installed:") && strings.Contains(line, "(") {
					// Extract size from installed line
					if start := strings.Index(line, "("); start != -1 {
						if end := strings.Index(line[start:], ")"); end != -1 {
							size = line[start+1 : start+end]
						}
					}
				}
			}
			
			packages = append(packages, pkg.Package{
				Name:        fields[0],
				Version:     fields[1],
				Description: description,
				Size:        size,
				Manager:     pkg.Snap,
			})
		}
	}

	return packages, scanner.Err()
}

func (s *SnapDetector) Uninstall(packageName string) error {
	cmd := exec.Command("pkexec", "snap", "remove", packageName)
	return cmd.Run()
}