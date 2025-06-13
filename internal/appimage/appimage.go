package appimage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"uninn/pkg"
)

type AppImageDetector struct{}

func New() *AppImageDetector {
	return &AppImageDetector{}
}

func (a *AppImageDetector) IsAvailable() bool {
	return true // AppImages are just files, always available
}

func (a *AppImageDetector) ListPackages() ([]pkg.Package, error) {
	var packages []pkg.Package
	
	// Common locations for AppImages
	searchPaths := []string{
		filepath.Join(os.Getenv("HOME"), "Applications"),
		filepath.Join(os.Getenv("HOME"), ".local/bin"),
		filepath.Join(os.Getenv("HOME"), "Downloads"),
		filepath.Join(os.Getenv("HOME"), "Desktop"),
		"/opt",
		"/usr/local/bin",
	}
	
	for _, path := range searchPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue walking
			}
			
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".appimage") {
				// Check if file is executable
				if info.Mode()&0111 != 0 {
					packages = append(packages, pkg.Package{
						Name:        strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
						Version:     "Unknown",
						Description: fmt.Sprintf("AppImage at %s", filepath.Dir(filePath)),
						Size:        fmt.Sprintf("%.2f MB", float64(info.Size())/(1024*1024)),
						Manager:     pkg.AppImage,
						Path:        filePath,
					})
				}
			}
			return nil
		})
		
		if err != nil {
			// Continue with other paths
			continue
		}
	}
	
	return packages, nil
}

func (a *AppImageDetector) Uninstall(packageName string) error {
	// For AppImages, we need the full path which is stored in the Path field
	// This should be handled by the caller
	return fmt.Errorf("use os.Remove with the full path to remove AppImage")
}