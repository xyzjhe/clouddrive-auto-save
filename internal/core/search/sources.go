// internal/core/search/sources.go
package search

import (
	"time"
)

// Source 搜索源接口
type Source interface {
	Name() string
	Search(query string, platforms []string, page int) (*SearchResult, error)
}

// SearchResult 搜索结果
type SearchResult struct {
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	Items      []SearchItem `json:"items"`
	NextCursor string       `json:"next_cursor,omitempty"`
}

// SearchItem 搜索结果项
type SearchItem struct {
	Title     string   `json:"title" binding:"required"`
	Source    string   `json:"source" binding:"required"`
	Platform  string   `json:"platform" binding:"required"`
	URL       string   `json:"url" binding:"required"`
	Summary   string   `json:"summary"`
	UpdatedAt string   `json:"updated_at"`
	Tags      []string `json:"tags,omitempty"`
	Channel   string   `json:"channel,omitempty"`
}

// cstLocation 中国标准时间 UTC+8
var cstLocation = func() *time.Location {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if loc == nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return loc
}()

// toCST 将 ISO 时间字符串转换为 CST 格式 YYYY-MM-DD HH:MM:SS
func toCST(isoTime string) string {
	if isoTime == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05-07:00", isoTime)
		if err != nil {
			return isoTime
		}
	}
	cst := t.In(cstLocation)
	if cst.Year() < 1970 {
		return ""
	}
	return cst.Format("2006-01-02 15:04:05")
}
