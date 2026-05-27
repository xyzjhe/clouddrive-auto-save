package renamer

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// MagicVariables 预定义的魔法正则变量
// 注意：Go 的 regexp 不支持 Lookaround (断言)，需使用 \b 或捕获组
var MagicVariables = map[string]string{
	"{YEAR}":    `\b(?:18|19|20)\d{2}\b`,
	"{DATE}":    `\b(?:18|19|20)?\d{2}[\.\-/年]\d{1,2}[\.\-/月]\d{1,2}\b`,
	"{CHINESE}": `\p{Han}{2,}`,
	"{EXT}":     `\.(\w+)$`, // 使用捕获组提取后缀
}

// 预编译全局静态正则表达式，消除运行期的高频 Compile 开销
var magicRegexps = map[string]*regexp.Regexp{
	"{YEAR}":    regexp.MustCompile(`\b(?:18|19|20)\d{2}\b`),
	"{DATE}":    regexp.MustCompile(`\b(?:18|19|20)?\d{2}[\.\-/年]\d{1,2}[\.\-/月]\d{1,2}\b`),
	"{CHINESE}": regexp.MustCompile(`\p{Han}{2,}`),
	"{EXT}":     regexp.MustCompile(`\.(\w+)$`),
}

var nonDigitRegexp = regexp.MustCompile(`\D`)

// PredefinedPattern 预定义魔法匹配规则
type PredefinedPattern struct {
	Pattern     string // 正则表达式
	Replacement string // 替换模板
	Description string // 说明
}

// PredefinedPatterns 预定义的命名规则集合
var PredefinedPatterns = map[string]PredefinedPattern{
	"$TV": {
		Pattern:     `(?i).*?([Ss]\d{1,2})?(?:[第EePpXx\.\-\_\( ]{1,2}|^)(\d{1,3})(?!\d).*?\.(mp4|mkv)`,
		Replacement: `$1E$2.$3`,
		Description: "剧集标准化命名 (S01E01.mp4)",
	},
	"$BLACK_WORD": {
		Pattern:     `^(?!.*纯享)(?!.*加更)(?!.*超前企划)(?!.*训练室)(?!.*蒸蒸日上).*$`,
		Replacement: `$0`,
		Description: "黑名单过滤 (排除纯享/加更/超前企划等)",
	},
	"$SHOW_MAGIC": {
		Pattern:     `^(?!.*纯享)(?!.*加更)(?!.*抢先)(?!.*预告).*?第\d+期.*`,
		Replacement: `{TASKNAME}.{SXX}E{II}.第{E}期{PART}.{EXT}`,
		Description: "综艺命名 (第X期格式)",
	},
	"$TV_MAGIC": {
		Pattern:     `.*\.(mp4|mkv|mov|m4v|avi|mpeg|ts)$`,
		Replacement: `{TASKNAME}.{SXX}E{E}.{EXT}`,
		Description: "通用视频命名 (剧集格式)",
	},
}

// GetPredefinedPattern 获取预定义匹配规则，如果不是预定义名称则返回 nil
func GetPredefinedPattern(name string) *PredefinedPattern {
	if p, ok := PredefinedPatterns[name]; ok {
		return &p
	}
	return nil
}

// RenameOptions 重命名选项
type RenameOptions struct {
	TaskName        string
	Pattern         string         // 用户定义的原始正则匹配式
	Replacement     string         // 用户定义的替换模板 (含变量或 Go template)
	FileName        string         // 原始文件名
	CompiledPattern *regexp.Regexp // 已编译的过滤正则表达式 (可选，复用以减少编译损耗)
}

// Processor 重命名处理器
type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

// Process 执行重命名逻辑
func (p *Processor) Process(opts RenameOptions) (string, error) {
	if opts.Replacement == "" {
		return opts.FileName, nil
	}

	result := opts.Replacement

	// 1. 替换基础变量 {TASKNAME} 和 {OLDNAME}
	result = strings.ReplaceAll(result, "{TASKNAME}", opts.TaskName)
	result = strings.ReplaceAll(result, "{OLDNAME}", opts.FileName)

	// 2. 尝试从原文件名中通过正则提取魔法变量的值并替换到 result 中
	for varName, re := range magicRegexps {
		if strings.Contains(result, varName) {
			matches := re.FindStringSubmatch(opts.FileName)
			if len(matches) > 0 {
				// 如果正则中有捕获组（如 {EXT}），则取第一个捕获组的内容
				// 否则取整个匹配到的字符串内容
				match := matches[0]
				if len(matches) > 1 {
					match = matches[1]
				}

				// 特殊处理日期格式
				if varName == "{DATE}" {
					match = p.cleanDate(match)
				}
				result = strings.ReplaceAll(result, varName, match)
			} else {
				// 未匹配到则置空
				result = strings.ReplaceAll(result, varName, "")
			}
		}
	}

	// 3. 执行正则子组替换 (如果 Pattern 和 Replacement 同时存在)
	if (opts.Pattern != "" || opts.CompiledPattern != nil) && strings.Contains(result, "$") {
		var re *regexp.Regexp
		var err error
		if opts.CompiledPattern != nil {
			re = opts.CompiledPattern
		} else {
			re, err = regexp.Compile(opts.Pattern)
		}
		if err == nil && re != nil {
			result = re.ReplaceAllString(opts.FileName, result)
		}
	}

	// 4. 执行 Go Template 动态渲染 (高级模式)
	if strings.Contains(result, "{{") {
		tmpl, err := template.New("rename").Parse(result)
		if err == nil {
			var buf bytes.Buffer
			data := map[string]interface{}{
				"TaskName": opts.TaskName,
				"OldName":  opts.FileName,
				"Now":      time.Now(),
			}
			if err := tmpl.Execute(&buf, data); err == nil {
				result = buf.String()
			}
		}
	}

	return strings.TrimSpace(result), nil
}

func (p *Processor) cleanDate(input string) string {
	// 移除非数字字符，统一为 YYYYMMDD 或 YYMMDD
	return nonDigitRegexp.ReplaceAllString(input, "")
}
