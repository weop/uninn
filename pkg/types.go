package pkg

type PackageManager string

const (
	APT       PackageManager = "apt"
	Pacman    PackageManager = "pacman"
	Flatpak   PackageManager = "flatpak"
	Snap      PackageManager = "snap"
	AppImage  PackageManager = "appimage"
	RPM       PackageManager = "rpm"
)

type Package struct {
	Name        string
	Version     string
	Description string
	Size        string
	Manager     PackageManager
	Path        string // For AppImages
}

type PackageDetector interface {
	ListPackages() ([]Package, error)
	Uninstall(packageName string) error
	IsAvailable() bool
}