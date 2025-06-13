package pacman

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"uninn/pkg"
)

type PacmanDetector struct{}

func New() *PacmanDetector {
	return &PacmanDetector{}
}

func (p *PacmanDetector) IsAvailable() bool {
	_, err := exec.LookPath("pacman")
	return err == nil
}

func (p *PacmanDetector) ListPackages() ([]pkg.Package, error) {
	cmd := exec.Command("pacman", "-Qi")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list pacman packages: %w", err)
	}

	var packages []pkg.Package
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	var currentPackage pkg.Package
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name            : ") {
			if currentPackage.Name != "" {
				packages = append(packages, currentPackage)
			}
			currentPackage = pkg.Package{
				Name:    strings.TrimPrefix(line, "Name            : "),
				Manager: pkg.Pacman,
			}
		} else if strings.HasPrefix(line, "Version         : ") {
			currentPackage.Version = strings.TrimPrefix(line, "Version         : ")
		} else if strings.HasPrefix(line, "Description     : ") {
			currentPackage.Description = strings.TrimPrefix(line, "Description     : ")
		} else if strings.HasPrefix(line, "Installed Size  : ") {
			currentPackage.Size = strings.TrimPrefix(line, "Installed Size  : ")
		}
	}
	
	if currentPackage.Name != "" {
		packages = append(packages, currentPackage)
	}

	return packages, scanner.Err()
}

func (p *PacmanDetector) Uninstall(packageName string) error {
	cmd := exec.Command("pkexec", "pacman", "-R", "--noconfirm", packageName)
	return cmd.Run()
}