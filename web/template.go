package web

// IndexTemplate struct to fill the index template
type IndexTemplate struct {
	State    string
	Log      []string
	Starting bool
	Online   bool
	Offline  bool
	Prefix   string
}
