package apt

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"uninn/pkg"
)

type APTDetector struct{}

func New() *APTDetector {
	return &APTDetector{}
}

func (a *APTDetector) IsAvailable() bool {
	_, err := exec.LookPath("dpkg")
	return err == nil
}

func (a *APTDetector) ListPackages() ([]pkg.Package, error) {
	cmd := exec.Command("dpkg-query", "-W", "-f=${Package}\t${Version}\t${binary:Summary}\t${Installed-Size}\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list apt packages: %w", err)
	}

	var packages []pkg.Package
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) >= 4 {
			packages = append(packages, pkg.Package{
				Name:        parts[0],
				Version:     parts[1],
				Description: parts[2],
				Size:        fmt.Sprintf("%s KB", parts[3]),
				Manager:     pkg.APT,
			})
		}
	}

	return packages, scanner.Err()
}

func (a *APTDetector) Uninstall(packageName string) error {
	cmd := exec.Command("pkexec", "apt-get", "remove", "-y", packageName)
	return cmd.Run()
}