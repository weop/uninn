package flatpak

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"uninn/pkg"
)

type FlatpakDetector struct{}

func New() *FlatpakDetector {
	return &FlatpakDetector{}
}

func (f *FlatpakDetector) IsAvailable() bool {
	_, err := exec.LookPath("flatpak")
	return err == nil
}

func (f *FlatpakDetector) ListPackages() ([]pkg.Package, error) {
	cmd := exec.Command("flatpak", "list", "--app", "--columns=application,version,description,size")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list flatpak packages: %w", err)
	}

	var packages []pkg.Package
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	// Skip header line
	if scanner.Scan() {
		scanner.Text()
	}
	
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 1 {
			// Get detailed info for each app
			infoCmd := exec.Command("flatpak", "info", parts[0])
			infoOutput, _ := infoCmd.Output()
			
			var version, size string
			infoScanner := bufio.NewScanner(strings.NewReader(string(infoOutput)))
			for infoScanner.Scan() {
				line := infoScanner.Text()
				if strings.Contains(line, "Version:") {
					version = strings.TrimSpace(strings.Split(line, ":")[1])
				} else if strings.Contains(line, "Installed size:") {
					size = strings.TrimSpace(strings.Split(line, ":")[1])
				}
			}
			
			packages = append(packages, pkg.Package{
				Name:        parts[0],
				Version:     version,
				Description: "Flatpak application",
				Size:        size,
				Manager:     pkg.Flatpak,
			})
		}
	}

	return packages, scanner.Err()
}

func (f *FlatpakDetector) Uninstall(packageName string) error {
	cmd := exec.Command("flatpak", "uninstall", "-y", packageName)
	return cmd.Run()
}