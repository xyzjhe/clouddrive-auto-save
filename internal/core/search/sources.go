// internal/core/search/sources.go
package search

// Source 搜索源接口
type Source interface {
	Name() string
	Search(query string, page int) (*SearchResult, error)
}

// SearchResult 搜索结果
type SearchResult struct {
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Items []SearchItem `json:"items"`
}

// SearchItem 搜索结果项
type SearchItem struct {
	Title     string `json:"title"`
	Source    string `json:"source"`
	Platform  string `json:"platform"`
	URL       string `json:"url"`
	Size      string `json:"size"`
	UpdatedAt string `json:"updated_at"`
	Summary   string `json:"summary"`
}

// CloudSaverSource CloudSaver 搜索源
type CloudSaverSource struct {
	baseURL string
}

func NewCloudSaverSource(baseURL string) *CloudSaverSource {
	return &CloudSaverSource{baseURL: baseURL}
}

func (s *CloudSaverSource) Name() string {
	return "CloudSaver"
}

func (s *CloudSaverSource) Search(query string, page int) (*SearchResult, error) {
	// TODO: 实现 CloudSaver 搜索
	return &SearchResult{}, nil
}

// PanSouSource PanSou 搜索源
type PanSouSource struct {
	baseURL string
}

func NewPanSouSource(baseURL string) *PanSouSource {
	return &PanSouSource{baseURL: baseURL}
}

func (s *PanSouSource) Name() string {
	return "PanSou"
}

func (s *PanSouSource) Search(query string, page int) (*SearchResult, error) {
	// TODO: 实现 PanSou 搜索
	return &SearchResult{}, nil
}
