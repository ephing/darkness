package chiho

import (
	"github.com/thecsw/darkness/emilia/alpha"
	"github.com/thecsw/darkness/emilia/alpha/roxy"
	"github.com/thecsw/darkness/emilia/narumi"
	"github.com/thecsw/darkness/emilia/puck"
	"github.com/thecsw/darkness/yunyun"
)

// EnrichPage enriches the page with the following:
// - Resolved comments
// - Enriched headings
// - Footnotes
// - Math support
// - Source code trimmed left whitespace
// - Syntax highlighting
// - Lazy galleries
// - Plugins
func EnrichPage(conf *alpha.DarknessConfig, page *yunyun.Page) *yunyun.Page {
	defer puck.Stopwatch("Enriched", "page", page.File).Record()
	options := []yunyun.PageOption{
		narumi.WithResolvedComments(),
		narumi.WithEnrichedHeadings(),
		narumi.WithFootnotes(),
		narumi.WithMathSupport(),
		narumi.WithSourceCodeTrimmedLeftWhitespace(),
		narumi.WithSyntaxHighlighting(conf),
		narumi.WithLazyGalleries(conf),
	}
	options = append(options, roxy.FormatForChiho(conf.Runtime.PluginConfigs, conf)...)
	return page.Options(options...)
}
