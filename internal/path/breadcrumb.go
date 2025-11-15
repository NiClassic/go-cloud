package path

type Breadcrumb struct {
	Name    string
	Path    string // DB path format
	URLPath string // URL format for links
	IsLast  bool
}
