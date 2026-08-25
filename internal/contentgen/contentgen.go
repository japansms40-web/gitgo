// Package contentgen 按模板和词库生成 Markdown 草稿。
//
// 这里是纯计算：随机源与当前时间都由调用方注入，不读写任何文件，
// 因此生成结果可以在测试里完全复现。文件读写在 internal/contentstore。
package contentgen

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// VarBankCount 是变量词库的数量，对应 {变量1}…{变量5}。
const VarBankCount = 5

// BodyTemplateCount 是正文模板的套数。生成每篇时随机选一套非空的，
// 让同一批草稿在结构上也有变化。
const BodyTemplateCount = 2

// MaxCount 是单次生成的篇数上限，防止手滑输入超大数字卡死界面。
const MaxCount = 500

// maxRandomLen 是 {英文=N} 这类随机串的长度上限。
const maxRandomLen = 64

// 关键词调用方式。
const (
	OrderSequential = "sequential" // 按词库顺序轮转
	OrderRandom     = "random"     // 每篇随机抽
)

// 关键词处理方式。
const (
	TransformNone  = "none"  // 原样使用
	TransformSpace = "space" // 字符之间插入空格
)

// Article 是文章库里的一篇现成素材，来自「文章库」目录下的一个文件。
type Article struct {
	Name string `json:"name"` // 文件名去掉扩展名，对应 {文章名}
	Body string `json:"body"` // 文件内容，对应 {文章}
}

// Library 是生成所需的全部素材，对应磁盘上的一组 txt 文件。
type Library struct {
	TitleTemplate string     `json:"titleTemplate"` // 标题模板，含占位符
	BodyTemplates []string   `json:"bodyTemplates"` // 正文模板 A/B，生成时随机选一套非空的
	Keywords      []string   `json:"keywords"`      // 关键词库，一行一条
	Vars          [][]string `json:"vars"`          // 变量词库，最多 VarBankCount 组，一行一条
	Images        []string   `json:"images"`        // 图片地址库，一行一条，对应 {图片}
	Articles      []Article  `json:"articles"`      // 文章库，对应 {文章} / {文章名}
}

// Options 控制生成行为，同时也是持久化到配置文件里的内容。
type Options struct {
	Count             int    `json:"count"`             // 本次生成篇数
	KeywordOrder      string `json:"keywordOrder"`      // OrderSequential | OrderRandom
	KeywordTransform  string `json:"keywordTransform"`  // TransformNone | TransformSpace
	ShuffleParagraphs bool   `json:"shuffleParagraphs"` // 正文段落随机排序
	DedupeLines       bool   `json:"dedupeLines"`       // 正文去除重复行
	ChineseOnly       bool   `json:"chineseOnly"`       // 正文仅保留含中文的行
}

// Normalize 把非法值纠正成可用的默认值。
func (o *Options) Normalize() {
	if o.Count < 1 {
		o.Count = 1
	}
	if o.Count > MaxCount {
		o.Count = MaxCount
	}
	if o.KeywordOrder != OrderRandom {
		o.KeywordOrder = OrderSequential
	}
	if o.KeywordTransform != TransformSpace {
		o.KeywordTransform = TransformNone
	}
}

// Draft 是一篇生成好的草稿。
type Draft struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// tokenRE 匹配 {…} 占位符。花括号是单字节 ASCII，所以按字节切片取名字是安全的。
var tokenRE = regexp.MustCompile(`\{[^{}]+\}`)

// 随机串的字符集。用 []rune 存是为了让 {中文=N} 这种多字节字符集也能按"个"取。
var charsets = map[string][]rune{
	"英文": []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"),
	"大写": []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"),
	"小写": []rune("abcdefghijklmnopqrstuvwxyz"),
	"数字": []rune("0123456789"),
	"字符": []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
	"中文": []rune("的一是了我不人在他有这个上们来到时大地为子中你说生国年着就那和要她出也得里后自以会家可下而过天去能对小多然于心学么之都好看起发当没成只如事把还用第样道想作种开美总从无情已面最女但现前些所同日手又行意动方期它头经长儿回位分爱老因很给名法间斯知世什两次使身者被高亲其进此话常与活正感"),
}

// {日期N} / {时间N} 的可选格式。空后缀等价于 1，方便直接写 {日期}。
var (
	dateLayouts = map[string]string{
		"": "2006-01-02", "1": "2006-01-02", "2": "2006/01/02",
		"3": "2006年01月02日", "4": "20060102",
	}
	timeLayouts = map[string]string{
		"": "15:04", "1": "15:04:05", "2": "15:04",
		"3": "15时04分", "4": "150405",
	}
)

// Generate 按 opts 从 lib 生成若干篇草稿。
//
// 词库为空时对应的占位符替换成空串（而不是报错），方便边填边预览；
// 只有标题和正文模板都为空时才没什么可生成的，返回错误。
func Generate(lib Library, opts Options, rnd *rand.Rand, now time.Time) ([]Draft, error) {
	opts.Normalize()

	bodies := nonEmpty(lib.BodyTemplates)
	if strings.TrimSpace(lib.TitleTemplate) == "" && len(bodies) == 0 {
		return nil, errors.New("标题模板和正文模板都为空，没有可生成的内容")
	}

	drafts := make([]Draft, 0, opts.Count)
	for i := 0; i < opts.Count; i++ {
		// 每篇先定好关键词和取哪一篇文章，这样 {文章名} 和 {文章}
		// 在同一篇草稿里指的是同一份素材。
		draw := drawContext{keyword: pickKeyword(lib.Keywords, i, opts, rnd)}
		if len(lib.Articles) > 0 {
			draw.article = lib.Articles[rnd.Intn(len(lib.Articles))]
		}

		title := strings.TrimSpace(render(lib.TitleTemplate, draw, lib, rnd, now))
		if title == "" {
			title = fmt.Sprintf("草稿 %d", i+1)
		}

		var body string
		if len(bodies) > 0 {
			body = postProcess(render(bodies[rnd.Intn(len(bodies))], draw, lib, rnd, now), opts, rnd)
		}

		drafts = append(drafts, Draft{Title: title, Body: body})
	}
	return drafts, nil
}

// drawContext 是一篇草稿内固定不变的取值，避免同一篇里前后不一致。
type drawContext struct {
	keyword string
	article Article
}

// nonEmpty 挑出真正填了内容的模板；A/B 只填一个也能正常生成。
func nonEmpty(templates []string) []string {
	var out []string
	for _, t := range templates {
		if strings.TrimSpace(t) != "" {
			out = append(out, t)
		}
	}
	return out
}

// pickKeyword 取出第 i 篇要用的关键词，并按处理方式变形。
func pickKeyword(keywords []string, i int, opts Options, rnd *rand.Rand) string {
	if len(keywords) == 0 {
		return ""
	}
	var kw string
	if opts.KeywordOrder == OrderRandom {
		kw = keywords[rnd.Intn(len(keywords))]
	} else {
		kw = keywords[i%len(keywords)]
	}
	if opts.KeywordTransform == TransformSpace {
		kw = strings.Join(strings.Split(kw, ""), " ")
	}
	return kw
}

// render 把模板里认识的占位符替换掉；不认识的原样留着，方便用户看出自己写错了。
func render(tpl string, draw drawContext, lib Library, rnd *rand.Rand, now time.Time) string {
	return tokenRE.ReplaceAllStringFunc(tpl, func(match string) string {
		if value, ok := resolve(match[1:len(match)-1], draw, lib, rnd, now); ok {
			return value
		}
		return match
	})
}

// resolve 解析单个占位符的名字，返回替换值；第二个返回值为 false 表示不认识这个占位符。
func resolve(name string, draw drawContext, lib Library, rnd *rand.Rand, now time.Time) (string, bool) {
	switch name {
	case "关键词":
		return draw.keyword, true
	case "文章":
		return draw.article.Body, true
	case "文章名":
		return draw.article.Name, true
	case "图片":
		return imageMarkdown(randomLine(lib.Images, rnd)), true
	}

	// {日期N} / {时间N}
	if rest, ok := strings.CutPrefix(name, "日期"); ok {
		layout, ok := dateLayouts[rest]
		return now.Format(layout), ok
	}
	if rest, ok := strings.CutPrefix(name, "时间"); ok {
		layout, ok := timeLayouts[rest]
		return now.Format(layout), ok
	}

	// {变量N}。合法范围由 VarBankCount 决定，与实际装了几组词库无关：
	// 词库没填时这是个空串，而不是"不认识的占位符"。
	if rest, ok := strings.CutPrefix(name, "变量"); ok {
		n, err := strconv.Atoi(rest)
		if err != nil || n < 1 || n > VarBankCount {
			return "", false
		}
		if n > len(lib.Vars) {
			return "", true
		}
		return randomLine(lib.Vars[n-1], rnd), true
	}

	// {英文=N} / {大写=N} / {小写=N} / {数字=N} / {字符=N} / {中文=N}
	if kind, arg, ok := strings.Cut(name, "="); ok {
		charset, known := charsets[kind]
		if !known {
			return "", false
		}
		n, err := strconv.Atoi(arg)
		if err != nil || n < 1 || n > maxRandomLen {
			return "", false
		}
		return randomRunes(charset, n, rnd), true
	}

	return "", false
}

// randomLine 从词库里随机抽一行；词库为空时返回空串。
func randomLine(bank []string, rnd *rand.Rand) string {
	if len(bank) == 0 {
		return ""
	}
	return bank[rnd.Intn(len(bank))]
}

// imageMarkdown 把图片库里的一行包成 Markdown 图片语法。
// 库里本来就写成 Markdown 或 HTML 的话就原样用，避免套两层。
func imageMarkdown(src string) string {
	if src == "" {
		return ""
	}
	if strings.HasPrefix(src, "![") || strings.HasPrefix(src, "<") {
		return src
	}
	return "![](" + src + ")"
}

func randomRunes(charset []rune, n int, rnd *rand.Rand) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteRune(charset[rnd.Intn(len(charset))])
	}
	return b.String()
}

// postProcess 按开关做正文的后处理：先整段打乱，再逐行过滤。
func postProcess(body string, opts Options, rnd *rand.Rand) string {
	if opts.ShuffleParagraphs {
		body = shuffleParagraphs(body, rnd)
	}
	return filterLines(body, opts)
}

// shuffleParagraphs 以空行为界把正文切成段落并随机排序。
func shuffleParagraphs(body string, rnd *rand.Rand) string {
	var paragraphs []string
	var current []string
	for _, line := range strings.Split(body, "\n") {
		if strings.TrimSpace(line) == "" {
			if len(current) > 0 {
				paragraphs = append(paragraphs, strings.Join(current, "\n"))
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		paragraphs = append(paragraphs, strings.Join(current, "\n"))
	}
	if len(paragraphs) < 2 {
		return body
	}

	rnd.Shuffle(len(paragraphs), func(i, j int) {
		paragraphs[i], paragraphs[j] = paragraphs[j], paragraphs[i]
	})
	return strings.Join(paragraphs, "\n\n")
}

// filterLines 逐行去重 / 只留中文。
// 空行不参与，这样段落结构能保留下来。
func filterLines(body string, opts Options) string {
	if !opts.DedupeLines && !opts.ChineseOnly {
		return body
	}
	lines := strings.Split(body, "\n")
	out := make([]string, 0, len(lines))
	seen := make(map[string]bool, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			out = append(out, line)
			continue
		}
		if opts.ChineseOnly && !hasChinese(line) {
			continue
		}
		if opts.DedupeLines {
			if seen[line] {
				continue
			}
			seen[line] = true
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func hasChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
