package path

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Converter struct {
	dataRoot string // The root directory for all file storage (e.g., "./data")
}

func New(dataRoot string) *Converter {
	return &Converter{dataRoot: dataRoot}
}

// ToDBPath converts any input path to the format stored in the database.
// Removes username prefix, leading/trailing slashes.
//
// Example: "/bob/documents/file.pdf" -> "documents/file.pdf"
//
// Example: "bob/documents/file.pdf" -> "documents/file.pdf"
func (c *Converter) ToDBPath(username, inputPath string) string {
	cleaned := strings.TrimPrefix(inputPath, "/")
	cleaned = strings.TrimPrefix(cleaned, username)
	cleaned = strings.TrimPrefix(cleaned, "/")

	if cleaned == "." || cleaned == "/" {
		return ""
	}

	return strings.TrimSuffix(cleaned, "/")
}

// FromDBPath converts a database path to a full filesystem path.
//
// Example: "documents/file.pdf" -> "/data/bob/documents/file.pdf"
func (c *Converter) FromDBPath(username, dbPath string) string {
	if dbPath == "" {
		return filepath.Join(c.dataRoot, username)
	}
	return filepath.Join(c.dataRoot, username, dbPath)
}

// ToURLPath converts a database path to a URL path for web display.
//
// Example: "documents/subfolder" -> "/files/documents/subfolder"
//
// Example: "" -> "/files/"
func (c *Converter) ToURLPath(dbPath string) string {
	if dbPath == "" {
		return "/files/"
	}
	return path.Join("/files", dbPath)
}

// FromURLPath extracts the relative path from a URL path.
//
// Example: "/files/documents/subfolder" -> "documents/subfolder"
//
// Example: "/files/" -> ""
func (c *Converter) FromURLPath(urlPath string) string {
	cleaned := strings.TrimPrefix(urlPath, "/files")
	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = strings.TrimSuffix(cleaned, "/")

	if cleaned == "." || cleaned == "/" {
		return ""
	}
	return cleaned
}

// JoinDBPath joins database path segments safely.
//
// Example: ("documents", "invoices") -> "documents/invoices"
//
// Example: ("", "documents") -> "documents"
func (c *Converter) JoinDBPath(segments ...string) string {
	// Filter out empty segments
	var nonEmpty []string
	for _, seg := range segments {
		seg = strings.Trim(seg, "/")
		if seg != "" && seg != "." {
			nonEmpty = append(nonEmpty, seg)
		}
	}

	if len(nonEmpty) == 0 {
		return ""
	}

	return strings.Join(nonEmpty, "/")
}

// GetParentDBPath returns the parent path of a database path.
//
// Example: "documents/invoices/2024" -> "documents/invoices"
//
// Example: "documents" -> ""
func (c *Converter) GetParentDBPath(dbPath string) string {
	if dbPath == "" {
		return ""
	}

	parent := path.Dir(dbPath)
	if parent == "." || parent == "/" {
		return ""
	}

	return parent
}

// GetBaseName returns the last element of the path
//
// Example: "documents/invoices/file.pdf" -> "file.pdf"
//
// Example: "documents" -> "documents"
func (c *Converter) GetBaseName(dbPath string) string {
	if dbPath == "" {
		return ""
	}
	return path.Base(dbPath)
}

// IsValidPath checks if a path is valid (no path traversal, etc.)
func (c *Converter) IsValidPath(inputPath string) bool {
	// Check for path traversal attempts
	if strings.Contains(inputPath, "..") {
		return false
	}

	// Check for absolute paths (we don't want those in DB)
	if strings.HasPrefix(inputPath, "/") && len(inputPath) > 1 {
		// Root "/" is ok, but "/something" should be relative
		return false
	}

	// Check for invalid characters
	invalidChars := []string{"\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(inputPath, char) {
			return false
		}
	}

	return true
}

// SanitizePath cleans and validates a path, returning an error if invalid
func (c *Converter) SanitizePath(inputPath string) (string, error) {
	if !c.IsValidPath(inputPath) {
		return "", fmt.Errorf("invalid path: %s", inputPath)
	}

	return path.Clean(inputPath), nil
}

// GetBreadcrumbs generates breadcrumbs from a database path.
//
// Example: "documents/invoices/2024" returns:
// [
//
//	{Name: "Home", Path: "", URLPath: "/files/", IsLast: false},
//	{Name: "documents", Path: "documents", URLPath: "/files/documents", IsLast: false},
//	{Name: "invoices", Path: "documents/invoices", URLPath: "/files/documents/invoices", IsLast: false},
//	{Name: "2024", Path: "documents/invoices/2024", URLPath: "/files/documents/invoices/2024", IsLast: true}
//
// ]
func (c *Converter) GetBreadcrumbs(dbPath string) []Breadcrumb {
	breadcrumbs := []Breadcrumb{
		{
			Name:    "Home",
			Path:    "",
			URLPath: "/files/",
			IsLast:  dbPath == "",
		},
	}

	if dbPath == "" {
		return breadcrumbs
	}

	parts := strings.Split(dbPath, "/")
	currentPath := ""

	for i, part := range parts {
		if part == "" {
			continue
		}

		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = currentPath + "/" + part
		}

		breadcrumb := Breadcrumb{
			Name:    part,
			Path:    currentPath,
			URLPath: c.ToURLPath(currentPath),
			IsLast:  i == len(parts)-1,
		}

		breadcrumbs = append(breadcrumbs, breadcrumb)
	}

	return breadcrumbs
}

// EnsureDir ensures a directory exists on the filesystem.
func (c *Converter) EnsureDir(username, dbPath string) error {
	fullPath := c.FromDBPath(username, dbPath)
	return os.MkdirAll(fullPath, 0755)
}

// GetFullFilePath returns the complete filesystem path for a file.
//
// Example: ("bob", "documents/invoice.pdf") -> "/data/bob/documents/invoice.pdf"
func (c *Converter) GetFullFilePath(username, dbPath string) string {
	return c.FromDBPath(username, dbPath)
}

// MigrateOldPath converts old path format to new DB format.
// This is useful during migration from the old system.
//
// Example: "/bob/documents/file.pdf" -> "documents/file.pdf"
//
// Example: "bob/documents/file.pdf" -> "documents/file.pdf"
func (c *Converter) MigrateOldPath(username, oldPath string) string {
	return c.ToDBPath(username, oldPath)
}

// IsRootPath checks if the path represents the root directory
func (c *Converter) IsRootPath(dbPath string) bool {
	return dbPath == ""
}

// IsChildOf checks if childPath is a child of parentPath.
//
// Example: IsChildOf("documents/invoices", "documents") -> true
//
// Example: IsChildOf("documents", "documents") -> false (same path)
//
// Example: IsChildOf("other/path", "documents") -> false
func (c *Converter) IsChildOf(childPath, parentPath string) bool {
	if parentPath == "" {
		// Everything is a child of root, except root itself
		return childPath != ""
	}

	if childPath == "" || childPath == parentPath {
		return false
	}

	// Check if child starts with parent + "/"
	return strings.HasPrefix(childPath, parentPath+"/")
}

// GetDepth returns the depth of a path (number of levels).
//
// Example: "" -> 0, "documents" -> 1, "documents/invoices" -> 2
func (c *Converter) GetDepth(dbPath string) int {
	if dbPath == "" {
		return 0
	}

	return strings.Count(dbPath, "/") + 1
}
