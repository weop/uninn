package rpm

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"uninn/pkg"
)

type RPMDetector struct {
	uninstallCmd string
}

func New() *RPMDetector {
	return &RPMDetector{}
}

func (d *RPMDetector) IsAvailable() bool {
	// Check if rpm command exists
	_, err := exec.LookPath("rpm")
	if err != nil {
		return false
	}

	// Detect which package manager to use for uninstall
	if _, err := exec.LookPath("dnf"); err == nil {
		d.uninstallCmd = "dnf"
	} else if _, err := exec.LookPath("yum"); err == nil {
		d.uninstallCmd = "yum"
	} else {
		d.uninstallCmd = "rpm"
	}

	return true
}

func (d *RPMDetector) ListPackages() ([]pkg.Package, error) {
	// Use rpm -qa with format string to get all info in one line per package
	cmd := exec.Command("rpm", "-qa", "--queryformat", "%{NAME}|%{VERSION}-%{RELEASE}|%{SIZE}|%{SUMMARY}\n")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list RPM packages: %w", err)
	}

	var packages []pkg.Package
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")
		
		if len(parts) >= 4 {
			// Format size from bytes to human readable
			sizeStr := formatSize(parts[2])
			
			pkg := pkg.Package{
				Name:        parts[0],
				Version:     parts[1],
				Size:        sizeStr,
				Description: parts[3],
				Manager:     pkg.RPM,
			}
			packages = append(packages, pkg)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning RPM output: %w", err)
	}

	return packages, nil
}

func (d *RPMDetector) Uninstall(packageName string) error {
	var cmd *exec.Cmd
	
	switch d.uninstallCmd {
	case "dnf":
		cmd = exec.Command("pkexec", "dnf", "remove", "-y", packageName)
	case "yum":
		cmd = exec.Command("pkexec", "yum", "remove", "-y", packageName)
	default:
		// fallback to rpm -e
		cmd = exec.Command("pkexec", "rpm", "-e", packageName)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall package: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func formatSize(sizeBytes string) string {
	// Convert size from bytes to human readable format
	var size int64
	fmt.Sscanf(sizeBytes, "%d", &size)
	
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])
}