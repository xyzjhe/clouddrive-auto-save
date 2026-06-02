package cloud139

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseShareLink(t *testing.T) {
	c := &Cloud139{}

	tests := []struct {
		name       string
		input      string
		wantLinkID string
		wantPasswd string
		wantCaID   string
		wantErr    bool
	}{
		{
			name:       "yun.139 Fragment 格式",
			input:      "https://yun.139.com/shareweb/#/w/i/105CqFasD8oLm",
			wantLinkID: "105CqFasD8oLm",
			wantCaID:   "root",
		},
		{
			name:       "caiyun.139 裸 query ID",
			input:      "https://caiyun.139.com/m/i?105CqFasD8oLm",
			wantLinkID: "105CqFasD8oLm",
			wantCaID:   "root",
		},
		{
			name:       "caiyun.139 带 http",
			input:      "http://caiyun.139.com/m/i?AbCd1234Ef",
			wantLinkID: "AbCd1234Ef",
			wantCaID:   "root",
		},
		{
			name:       "cloud.139 带 linkID 参数",
			input:      "https://cloud.139.com/w/i/?linkID=XYZ789&pCaID=sub1",
			wantLinkID: "XYZ789",
			wantCaID:   "sub1",
		},
		{
			name:       "cloud.139 带 passwd 参数",
			input:      "https://cloud.139.com/w/i/?linkID=ABC123&passwd=9999",
			wantLinkID: "ABC123",
			wantPasswd: "9999",
			wantCaID:   "root",
		},
		{
			name:       "cloud.139 带 pwd 参数",
			input:      "https://cloud.139.com/w/i/?linkID=ABC123&pwd=abcd",
			wantLinkID: "ABC123",
			wantPasswd: "abcd",
			wantCaID:   "root",
		},
		{
			name:    "空链接",
			input:   "",
			wantErr: true,
		},
		{
			name:    "无效链接无 ID",
			input:   "https://cloud.139.com/w/",
			wantErr: true,
		},
		{
			name:       "裸 ID 直接输入",
			input:      "105CqFasD8oLm",
			wantLinkID: "105CqFasD8oLm",
			wantCaID:   "root",
		},
		{
			name:       "caiyun.139 裸 query ID 尾部带 emoji",
			input:      "https://caiyun.139.com/m/i?2j3ahPrfeGrp0🏷",
			wantLinkID: "2j3ahPrfeGrp0",
			wantCaID:   "root",
		},
		{
			name:       "yun.139 Fragment 格式尾部带特殊符号",
			input:      "https://yun.139.com/shareweb/#/w/i/ABC123!!!",
			wantLinkID: "ABC123",
			wantCaID:   "root",
		},
		{
			name:       "Fragment 带 query 参数",
			input:      "https://yun.139.com/shareweb/#/w/i/ABC?linkID=fromFragment&pCaID=fold1",
			wantLinkID: "fromFragment",
			wantCaID:   "fold1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkID, passwd, pCaID, err := c.parseShareLink(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantLinkID, linkID)
			assert.Equal(t, tt.wantPasswd, passwd)
			assert.Equal(t, tt.wantCaID, pCaID)
		})
	}
}
