package contentgen

import (
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

// fixedNow 让 {日期N}/{时间N} 的断言可重复。
var fixedNow = time.Date(2026, 8, 25, 17, 42, 30, 0, time.UTC)

func newRnd() *rand.Rand { return rand.New(rand.NewSource(1)) }

// body 把单套正文模板包成 Library 需要的切片。
func body(tpl string) []string { return []string{tpl} }

func TestGenerateReplacesKnownTokens(t *testing.T) {
	lib := Library{
		TitleTemplate: "{关键词}-{日期}",
		BodyTemplates: body("写于 {时间}，关于 {关键词}。"),
		Keywords:      []string{"茶叶"},
	}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if len(drafts) != 1 {
		t.Fatalf("期望 1 篇，实际 %d 篇", len(drafts))
	}
	if got, want := drafts[0].Title, "茶叶-2026-08-25"; got != want {
		t.Errorf("标题 = %q，期望 %q", got, want)
	}
	if got, want := drafts[0].Body, "写于 17:42，关于 茶叶。"; got != want {
		t.Errorf("正文 = %q，期望 %q", got, want)
	}
}

func TestGenerateDateAndTimeVariants(t *testing.T) {
	cases := map[string]string{
		"{日期1}": "2026-08-25",
		"{日期2}": "2026/08/25",
		"{日期3}": "2026年08月25日",
		"{日期4}": "20260825",
		"{时间1}": "17:42:30",
		"{时间2}": "17:42",
		"{时间3}": "17时42分",
		"{时间4}": "174230",
		"{时间5}": "17时42分30秒",
	}
	for token, want := range cases {
		t.Run(token, func(t *testing.T) {
			lib := Library{TitleTemplate: token, BodyTemplates: body("x")}
			drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
			if err != nil {
				t.Fatalf("Generate 返回错误: %v", err)
			}
			if drafts[0].Title != want {
				t.Errorf("%s = %q，期望 %q", token, drafts[0].Title, want)
			}
		})
	}
}

func TestGenerateRejectsOutOfRangeDateTimeVariant(t *testing.T) {
	lib := Library{TitleTemplate: "标题", BodyTemplates: body("{日期9}{时间0}")}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Body, "{日期9}{时间0}"; got != want {
		t.Errorf("越界的日期/时间编号应原样保留，得到 %q", got)
	}
}

func TestGenerateKeepsUnknownToken(t *testing.T) {
	lib := Library{TitleTemplate: "{不存在的东西}", BodyTemplates: body("正文")}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Title, "{不存在的东西}"; got != want {
		t.Errorf("未知 token 应原样保留，得到 %q，期望 %q", got, want)
	}
}

func TestGenerateVarTokenDrawnIndependentlyPerOccurrence(t *testing.T) {
	bank := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	lib := Library{
		TitleTemplate: "标题",
		BodyTemplates: body("{变量1}{变量1}{变量1}{变量1}{变量1}{变量1}{变量1}{变量1}"),
		Vars:          [][]string{bank},
	}

	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	runes := []rune(drafts[0].Body)
	if len(runes) != 8 {
		t.Fatalf("期望 8 次替换，得到 %q", drafts[0].Body)
	}
	distinct := map[rune]bool{}
	for _, r := range runes {
		distinct[r] = true
		if !strings.ContainsRune(strings.Join(bank, ""), r) {
			t.Errorf("抽到了词库以外的内容 %q", string(r))
		}
	}
	// 每个 token 独立重抽，8 次不可能全部相同（种子固定，结果确定）。
	if len(distinct) < 2 {
		t.Errorf("同一 token 多次出现应各自独立抽取，实际全部相同: %q", drafts[0].Body)
	}
}

func TestGenerateSequentialKeywordsRotate(t *testing.T) {
	lib := Library{
		TitleTemplate: "{关键词}",
		BodyTemplates: body("x"),
		Keywords:      []string{"甲", "乙", "丙"},
	}
	drafts, err := Generate(lib, Options{Count: 5, KeywordOrder: OrderSequential}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	for i, want := range []string{"甲", "乙", "丙", "甲", "乙"} {
		if drafts[i].Title != want {
			t.Errorf("第 %d 篇标题 = %q，期望 %q", i, drafts[i].Title, want)
		}
	}
}

func TestGenerateRandomKeywordsStayInBank(t *testing.T) {
	bank := []string{"甲", "乙", "丙"}
	lib := Library{TitleTemplate: "{关键词}", BodyTemplates: body("x"), Keywords: bank}
	drafts, err := Generate(lib, Options{Count: 20, KeywordOrder: OrderRandom}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	for i, d := range drafts {
		if !slicesContains(bank, d.Title) {
			t.Errorf("第 %d 篇抽到了词库以外的关键词 %q", i, d.Title)
		}
	}
}

func TestGenerateKeywordTransformSpace(t *testing.T) {
	lib := Library{TitleTemplate: "{关键词}", BodyTemplates: body("x"), Keywords: []string{"茶叶"}}
	drafts, err := Generate(lib, Options{Count: 1, KeywordTransform: TransformSpace}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Title, "茶 叶"; got != want {
		t.Errorf("加空格后 = %q，期望 %q", got, want)
	}
}

func TestGenerateRandomStringTokens(t *testing.T) {
	cases := []struct {
		token   string
		want    int
		charset string
	}{
		{"{英文=6}", 6, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"{大写=5}", 5, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"{小写=4}", 4, "abcdefghijklmnopqrstuvwxyz"},
		{"{数字=3}", 3, "0123456789"},
		{"{字符=8}", 8, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"},
	}
	for _, tc := range cases {
		t.Run(tc.token, func(t *testing.T) {
			lib := Library{TitleTemplate: tc.token, BodyTemplates: body("x")}
			drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
			if err != nil {
				t.Fatalf("Generate 返回错误: %v", err)
			}
			got := drafts[0].Title
			if len([]rune(got)) != tc.want {
				t.Errorf("%s 应产生 %d 位，得到 %q", tc.token, tc.want, got)
			}
			if strings.Trim(got, tc.charset) != "" {
				t.Errorf("%s 产生了字符集以外的内容: %q", tc.token, got)
			}
		})
	}
}

func TestGenerateChineseRandomToken(t *testing.T) {
	lib := Library{TitleTemplate: "{中文=5}", BodyTemplates: body("x")}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	got := []rune(drafts[0].Title)
	if len(got) != 5 {
		t.Fatalf("{中文=5} 应产生 5 个字，得到 %q", drafts[0].Title)
	}
	for _, r := range got {
		if !hasChinese(string(r)) {
			t.Errorf("产生了非汉字 %q", string(r))
		}
	}
}

func TestGenerateImageToken(t *testing.T) {
	cases := []struct {
		name string
		bank []string
		want string
	}{
		{"裸地址包成 Markdown", []string{"https://example.com/a.png"}, "![](https://example.com/a.png)"},
		{"已是 Markdown 则原样用", []string{"![图](https://example.com/a.png)"}, "![图](https://example.com/a.png)"},
		{"已是 HTML 则原样用", []string{`<img src="a.png" />`}, `<img src="a.png" />`},
		{"图片库为空替换成空串", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lib := Library{TitleTemplate: "标题", BodyTemplates: body("{图片}"), Images: tc.bank}
			drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
			if err != nil {
				t.Fatalf("Generate 返回错误: %v", err)
			}
			if drafts[0].Body != tc.want {
				t.Errorf("{图片} = %q，期望 %q", drafts[0].Body, tc.want)
			}
		})
	}
}

func TestGenerateArticleTokens(t *testing.T) {
	lib := Library{
		TitleTemplate: "{文章名}",
		BodyTemplates: body("{文章}"),
		Articles:      []Article{{Name: "甲篇", Body: "甲的正文"}},
	}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if drafts[0].Title != "甲篇" {
		t.Errorf("{文章名} = %q，期望 %q", drafts[0].Title, "甲篇")
	}
	if drafts[0].Body != "甲的正文" {
		t.Errorf("{文章} = %q，期望 %q", drafts[0].Body, "甲的正文")
	}
}

// 同一篇草稿里 {文章名} 和 {文章} 必须来自同一份素材，
// 否则会出现"标题是甲篇、正文是乙篇"这种错位。
func TestGenerateArticleTokensStayConsistentWithinDraft(t *testing.T) {
	articles := []Article{
		{Name: "甲", Body: "甲的正文"},
		{Name: "乙", Body: "乙的正文"},
		{Name: "丙", Body: "丙的正文"},
	}
	lib := Library{
		TitleTemplate: "{文章名}",
		BodyTemplates: body("{文章}|{文章名}"),
		Articles:      articles,
	}
	drafts, err := Generate(lib, Options{Count: 30}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}

	used := map[string]bool{}
	for i, d := range drafts {
		want := d.Title + "的正文|" + d.Title
		if d.Body != want {
			t.Fatalf("第 %d 篇标题与正文错位：标题 %q，正文 %q", i, d.Title, d.Body)
		}
		used[d.Title] = true
	}
	if len(used) < 2 {
		t.Errorf("30 篇里应当抽到多篇不同文章，实际只用到 %v", used)
	}
}

func TestGenerateArticleTokensWithEmptyLibrary(t *testing.T) {
	lib := Library{TitleTemplate: "标题", BodyTemplates: body("[{文章}][{文章名}]")}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("文章库为空不应报错，得到: %v", err)
	}
	if got, want := drafts[0].Body, "[][]"; got != want {
		t.Errorf("文章库为空应替换成空串，得到 %q", got)
	}
}

func TestGenerateRejectsBadRandomLength(t *testing.T) {
	lib := Library{TitleTemplate: "标题", BodyTemplates: body("{英文=0}{小写=abc}{数字=999}")}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Body, "{英文=0}{小写=abc}{数字=999}"; got != want {
		t.Errorf("非法长度应原样保留，得到 %q", got)
	}
}

func TestGeneratePicksAmongBodyTemplates(t *testing.T) {
	lib := Library{
		TitleTemplate: "标题",
		BodyTemplates: []string{"模板A", "模板B"},
	}
	drafts, err := Generate(lib, Options{Count: 30}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	seen := map[string]bool{}
	for _, d := range drafts {
		if d.Body != "模板A" && d.Body != "模板B" {
			t.Fatalf("产生了预期外的正文: %q", d.Body)
		}
		seen[d.Body] = true
	}
	if len(seen) != 2 {
		t.Errorf("30 篇里应当两套模板都用到，实际只用到 %v", seen)
	}
}

func TestGenerateSkipsEmptyBodyTemplate(t *testing.T) {
	lib := Library{
		TitleTemplate: "标题",
		BodyTemplates: []string{"", "只有 B 有内容"},
	}
	drafts, err := Generate(lib, Options{Count: 5}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	for i, d := range drafts {
		if d.Body != "只有 B 有内容" {
			t.Errorf("第 %d 篇用到了空模板: %q", i, d.Body)
		}
	}
}

func TestGenerateShuffleParagraphs(t *testing.T) {
	tpl := "第一段\n第一段续行\n\n第二段\n\n第三段\n\n第四段\n\n第五段"
	lib := Library{TitleTemplate: "标题", BodyTemplates: body(tpl)}

	drafts, err := Generate(lib, Options{Count: 1, ShuffleParagraphs: true}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	got := drafts[0].Body

	if got == tpl {
		t.Errorf("段落应被打乱，结果与原文一致: %q", got)
	}
	// 打乱只是重排，内容一段都不能少，且段内的行必须黏在一起。
	for _, want := range []string{"第一段\n第一段续行", "第二段", "第三段", "第四段", "第五段"} {
		if !strings.Contains(got, want) {
			t.Errorf("打乱后丢了段落 %q，结果: %q", want, got)
		}
	}
}

func TestGenerateShuffleKeepsSingleParagraphIntact(t *testing.T) {
	tpl := "只有一段\n但是有两行"
	lib := Library{TitleTemplate: "标题", BodyTemplates: body(tpl)}
	drafts, err := Generate(lib, Options{Count: 1, ShuffleParagraphs: true}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if drafts[0].Body != tpl {
		t.Errorf("只有一段时不该改动，得到 %q", drafts[0].Body)
	}
}

func TestGenerateEmptyBanksAreLenient(t *testing.T) {
	lib := Library{TitleTemplate: "[{关键词}]", BodyTemplates: body("[{变量3}]")}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("词库为空不应报错，得到: %v", err)
	}
	if got, want := drafts[0].Title, "[]"; got != want {
		t.Errorf("空关键词库应替换为空串，得到 %q", got)
	}
	if got, want := drafts[0].Body, "[]"; got != want {
		t.Errorf("空变量库应替换为空串，得到 %q", got)
	}
}

func TestGenerateErrorsWhenAllTemplatesEmpty(t *testing.T) {
	if _, err := Generate(Library{BodyTemplates: []string{"", "  "}}, Options{Count: 1}, newRnd(), fixedNow); err == nil {
		t.Error("标题与两套正文模板都为空时应返回错误")
	}
}

func TestGenerateTitleFallbackWhenEmpty(t *testing.T) {
	lib := Library{TitleTemplate: "  ", BodyTemplates: body("正文")}
	drafts, err := Generate(lib, Options{Count: 2}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if drafts[0].Title == "" || drafts[1].Title == "" {
		t.Error("标题为空时应有兜底标题")
	}
	if drafts[0].Title == drafts[1].Title {
		t.Errorf("兜底标题应可区分，两篇都是 %q", drafts[0].Title)
	}
}

func TestGenerateDedupeLines(t *testing.T) {
	lib := Library{
		TitleTemplate: "标题",
		BodyTemplates: body("一\n二\n一\n\n\n三\n二"),
	}
	drafts, err := Generate(lib, Options{Count: 1, DedupeLines: true}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Body, "一\n二\n\n\n三"; got != want {
		t.Errorf("去重后 = %q，期望 %q（空行应保留，用于分段）", got, want)
	}
}

func TestGenerateChineseOnly(t *testing.T) {
	lib := Library{
		TitleTemplate: "标题",
		BodyTemplates: body("中文一行\npure english\n混合 mixed 中文\n12345\n\n结尾"),
	}
	drafts, err := Generate(lib, Options{Count: 1, ChineseOnly: true}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Body, "中文一行\n混合 mixed 中文\n\n结尾"; got != want {
		t.Errorf("仅保留中文行后 = %q，期望 %q", got, want)
	}
}

func TestOptionsNormalize(t *testing.T) {
	cases := []struct {
		name string
		in   Options
		want Options
	}{
		{
			name: "零值补默认",
			in:   Options{},
			want: Options{Count: 1, KeywordOrder: OrderSequential, KeywordTransform: TransformNone},
		},
		{
			name: "负数篇数归一",
			in:   Options{Count: -3},
			want: Options{Count: 1, KeywordOrder: OrderSequential, KeywordTransform: TransformNone},
		},
		{
			name: "超上限截断",
			in:   Options{Count: 99999},
			want: Options{Count: MaxCount, KeywordOrder: OrderSequential, KeywordTransform: TransformNone},
		},
		{
			name: "非法枚举兜底",
			in:   Options{Count: 3, KeywordOrder: "乱写", KeywordTransform: "乱写"},
			want: Options{Count: 3, KeywordOrder: OrderSequential, KeywordTransform: TransformNone},
		},
		{
			name: "合法值保持不变",
			in:   Options{Count: 7, KeywordOrder: OrderRandom, KeywordTransform: TransformSpace, DedupeLines: true, ShuffleParagraphs: true},
			want: Options{Count: 7, KeywordOrder: OrderRandom, KeywordTransform: TransformSpace, DedupeLines: true, ShuffleParagraphs: true},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in
			got.Normalize()
			if got != tc.want {
				t.Errorf("Normalize() = %+v，期望 %+v", got, tc.want)
			}
		})
	}
}

func TestGenerateRespectsCount(t *testing.T) {
	lib := Library{TitleTemplate: "标题", BodyTemplates: body("正文")}
	drafts, err := Generate(lib, Options{Count: 12}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if len(drafts) != 12 {
		t.Errorf("期望 12 篇，实际 %d 篇", len(drafts))
	}
}

func TestGenerateRejectsOutOfRangeVarToken(t *testing.T) {
	lib := Library{TitleTemplate: "标题", BodyTemplates: body("{变量0}{变量6}{变量99}")}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Body, "{变量0}{变量6}{变量99}"; got != want {
		t.Errorf("越界的变量 token 应原样保留，得到 %q", got)
	}
}

func slicesContains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// ---------- 循环标签 {循环=N}…{/循环} ----------

func TestExpandLoopsRepeatsBlock(t *testing.T) {
	if got := expandLoops("{循环=3}X{/循环}"); got != "XXX" {
		t.Errorf("expandLoops = %q，期望 XXX", got)
	}
}

func TestExpandLoopsFlatSiblings(t *testing.T) {
	// 平铺的兄弟循环应各自就近配对，互不干扰。
	if got := expandLoops("{循环=2}A{/循环}-{循环=3}B{/循环}"); got != "AA-BBB" {
		t.Errorf("expandLoops = %q，期望 AA-BBB", got)
	}
}

func TestExpandLoopsClampsUpperBound(t *testing.T) {
	got := expandLoops("{循环=99999}x{/循环}")
	if len(got) != maxLoopN {
		t.Errorf("超上限的循环应钳到 maxLoopN=%d，实际展开 %d 份", maxLoopN, len(got))
	}
}

func TestExpandLoopsVolumeCap(t *testing.T) {
	block := strings.Repeat("a", 10000)
	got := expandLoops("{循环=99999}" + block + "{/循环}")
	if len(got) > maxExpandedLen {
		t.Errorf("展开体积 %d 超过上限 %d", len(got), maxExpandedLen)
	}
}

func TestExpandLoopsMissingCloseKeptLiteral(t *testing.T) {
	in := "{循环=3}abc"
	if got := expandLoops(in); got != in {
		t.Errorf("缺 {/循环} 应原样保留，得到 %q", got)
	}
}

func TestExpandLoopsInvalidNKeptLiteral(t *testing.T) {
	in := "{循环=0}x{/循环}"
	if got := expandLoops(in); got != in {
		t.Errorf("{循环=0} 应原样保留，得到 %q", got)
	}
}

func TestGenerateLoopRandomIndependentPerIteration(t *testing.T) {
	bank := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	lib := Library{
		TitleTemplate: "标题",
		BodyTemplates: body("{循环=8}{变量1}{/循环}"),
		Vars:          [][]string{bank},
	}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	runes := []rune(drafts[0].Body)
	if len(runes) != 8 {
		t.Fatalf("循环 8 次应有 8 个字符，得到 %q", drafts[0].Body)
	}
	distinct := map[rune]bool{}
	for _, r := range runes {
		distinct[r] = true
	}
	if len(distinct) < 2 {
		t.Errorf("循环内的随机标签应各次独立重抽，实际全同: %q", drafts[0].Body)
	}
}

func TestGenerateLoopKeepsKeywordConstant(t *testing.T) {
	lib := Library{
		TitleTemplate: "{循环=3}{关键词}|{/循环}",
		BodyTemplates: body("x"),
		Keywords:      []string{"茶"},
	}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Title, "茶|茶|茶|"; got != want {
		t.Errorf("循环内 {关键词} 应保持同值，得到 %q，期望 %q", got, want)
	}
}

// ---------- 外链标签 {随机外链} / {顺序外链} ----------

func TestGenerateRandomExternalLink(t *testing.T) {
	urls := []string{"http://a.cn", "http://b.cn", "http://c.cn"}
	lib := Library{TitleTemplate: "{随机外链}", BodyTemplates: body("x"), URLs: urls}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if !slicesContains(urls, drafts[0].Title) {
		t.Errorf("{随机外链} 应取自外链库，得到 %q", drafts[0].Title)
	}
}

func TestGenerateSequentialExternalLinkRotates(t *testing.T) {
	// 用标题（不过 postProcess，只 TrimSpace）验证顺序轮转，完全确定、与 rnd 无关。
	lib := Library{
		TitleTemplate: "{循环=5}{顺序外链},{/循环}",
		BodyTemplates: body("x"),
		URLs:          []string{"u0", "u1", "u2"},
	}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Title, "u0,u1,u2,u0,u1,"; got != want {
		t.Errorf("{顺序外链} 应按行轮转，得到 %q，期望 %q", got, want)
	}
}

func TestGenerateExternalLinkEmptyBankLenient(t *testing.T) {
	lib := Library{TitleTemplate: "[{随机外链}][{顺序外链}]", BodyTemplates: body("x")}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Title, "[][]"; got != want {
		t.Errorf("空外链库应替换成空串，得到 %q，期望 %q", got, want)
	}
}

// ---------- {顺序外链} 全局顺序消费 ----------

// TestGenerateSequentialExternalLinkAcrossDrafts 一次 Generate 里的多篇必须接着往下取，
// 而不是每篇都从外链库第一条重新开始。
func TestGenerateSequentialExternalLinkAcrossDrafts(t *testing.T) {
	lib := Library{
		TitleTemplate: "{顺序外链}",
		BodyTemplates: body("x"),
		URLs:          []string{"u0", "u1", "u2", "u3", "u4"},
	}
	drafts, err := Generate(lib, Options{Count: 4}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	want := []string{"u0", "u1", "u2", "u3"}
	for i, w := range want {
		if got := drafts[i].Title; got != w {
			t.Errorf("第 %d 篇 {顺序外链} 得到 %q，期望 %q", i+1, got, w)
		}
	}
}

// TestGenerateSequentialExternalLinkTitleAndBodyShareCursor 同一篇里标题与正文共用一个游标，
// 不会各自从头取，保证一条链接只被用一次。
func TestGenerateSequentialExternalLinkTitleAndBodyShareCursor(t *testing.T) {
	lib := Library{
		TitleTemplate: "{顺序外链}",
		BodyTemplates: []string{"{顺序外链}"},
		URLs:          []string{"u0", "u1", "u2"},
	}
	drafts, err := Generate(lib, Options{Count: 1}, newRnd(), fixedNow)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if got, want := drafts[0].Title, "u0"; got != want {
		t.Errorf("标题得到 %q，期望 %q", got, want)
	}
	if got, want := drafts[0].Body, "u1"; got != want {
		t.Errorf("正文应接着标题往下取，得到 %q，期望 %q", got, want)
	}
}

// TestGenerateWithSharedCursorContinuesAcrossCalls 注入的共享游标要跨多次 GenerateWith
// 调用继续往下走——一次批量发布里换了账号也不会重复用前面的链接。
func TestGenerateWithSharedCursorContinuesAcrossCalls(t *testing.T) {
	lib := Library{
		TitleTemplate: "{顺序外链}",
		BodyTemplates: body("x"),
		URLs:          []string{"u0", "u1", "u2", "u3", "u4", "u5"},
	}
	seq := NewSeqCursor()
	var got []string
	for call := 0; call < 3; call++ {
		drafts, err := GenerateWith(lib, Options{Count: 2}, newRnd(), fixedNow, seq)
		if err != nil {
			t.Fatalf("第 %d 次 GenerateWith 返回错误: %v", call+1, err)
		}
		for _, d := range drafts {
			got = append(got, d.Title)
		}
	}
	want := []string{"u0", "u1", "u2", "u3", "u4", "u5"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("共享游标应跨调用连续消费，得到 %v，期望 %v", got, want)
	}
}

// TestGenerateWithSharedCursorWrapsAfterExhausted 全部用完一轮后才回到第一条。
func TestGenerateWithSharedCursorWrapsAfterExhausted(t *testing.T) {
	lib := Library{
		TitleTemplate: "{顺序外链}",
		BodyTemplates: body("x"),
		URLs:          []string{"u0", "u1", "u2"},
	}
	seq := NewSeqCursor()
	drafts, err := GenerateWith(lib, Options{Count: 5}, newRnd(), fixedNow, seq)
	if err != nil {
		t.Fatalf("GenerateWith 返回错误: %v", err)
	}
	want := []string{"u0", "u1", "u2", "u0", "u1"}
	for i, w := range want {
		if got := drafts[i].Title; got != w {
			t.Errorf("第 %d 篇得到 %q，期望 %q", i+1, got, w)
		}
	}
}

// TestGenerateWithNilCursorFallsBack 传 nil 游标要退化成"本次调用内共享"，不能 panic。
func TestGenerateWithNilCursorFallsBack(t *testing.T) {
	lib := Library{
		TitleTemplate: "{顺序外链}",
		BodyTemplates: body("x"),
		URLs:          []string{"u0", "u1", "u2"},
	}
	drafts, err := GenerateWith(lib, Options{Count: 2}, newRnd(), fixedNow, nil)
	if err != nil {
		t.Fatalf("GenerateWith 返回错误: %v", err)
	}
	if got, want := drafts[1].Title, "u1"; got != want {
		t.Errorf("nil 游标应退化成本次调用内共享，得到 %q，期望 %q", got, want)
	}
}

// TestSeqCursorConcurrentUnique 多 worker 并发取号不重复（配 -race 跑）。
func TestSeqCursorConcurrentUnique(t *testing.T) {
	const workers, perWorker = 8, 500
	seq := NewSeqCursor()
	out := make(chan int64, workers*perWorker)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				out <- seq.Next()
			}
		}()
	}
	wg.Wait()
	close(out)

	seen := make(map[int64]bool, workers*perWorker)
	for n := range out {
		if seen[n] {
			t.Fatalf("序号 %d 被取到两次", n)
		}
		seen[n] = true
	}
	if len(seen) != workers*perWorker {
		t.Errorf("取到 %d 个不同序号，期望 %d 个", len(seen), workers*perWorker)
	}
}

// TestSeqCursorNilSafe 零值/nil 游标都不能 panic。
func TestSeqCursorNilSafe(t *testing.T) {
	var zero SeqCursor
	if got := zero.Next(); got != 0 {
		t.Errorf("零值游标首次 Next 应为 0，得到 %d", got)
	}
	var nilCur *SeqCursor
	if got := nilCur.Next(); got != 0 {
		t.Errorf("nil 游标 Next 应为 0，得到 %d", got)
	}
}
