// internal/core/search/stubs.go
// 占位桩函数，将在 Task 3/4 中被真实实现替换
package search

// NewCloudSaverSource 创建 CloudSaver 搜索源（占位桩，待 Task 3 实现）
func NewCloudSaverSource(server string) Source {
	return &stubSource{name: "cloudsaver"}
}

// NewPanSouSource 创建 PanSou 搜索源（占位桩，待 Task 4 实现）
func NewPanSouSource(server string) Source {
	return &stubSource{name: "pansou"}
}

// stubSource 占位搜索源实现
type stubSource struct {
	name string
}

func (s *stubSource) Name() string { return s.name }

func (s *stubSource) Search(query string, page int) (*SearchResult, error) {
	return &SearchResult{}, nil
}
