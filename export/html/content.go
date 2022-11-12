package html

import (
	"fmt"
	"html"
	"strings"

	"github.com/thecsw/darkness/yunyun"
	"github.com/thecsw/gana"
)

// buildContent runs the givent `*Content` against known protocols/policies
// and does some funky logic to balance div openings and closures.
func (e *ExporterHTML) buildContent(i int, v *yunyun.Content) string {
	built := e.contentFunctions[v.Type](v)
	divv := e.whatDivType(v)
	if e.inHeading {
		if i == e.contentsNum-1 {
			built += "</div>\n</div>"
			goto done
		}
		if divv != divOutside {
			goto done
		}
		built = "</div>\n</div>" + built
		e.inHeading = false
	}
done:
	if e.inWriting {
		if divv != divOutside {
			e.inWriting = divv == divWriting
			goto done2
		}
		built = "</div>" + built
		e.inWriting = false
	} else {
		if divv == divWriting {
			built = `<div class="writing">` + built
			e.inWriting = true
		}
	}
done2:
	if divv == divWriting && i == e.contentsNum-1 {
		built += "</div>"
	}
	return built
}

// headings gives us a heading html representation
func (e *ExporterHTML) Heading(content *yunyun.Content) string {
	start := ``
	// Automatically close whatever you had before
	if e.inHeading {
		start = "</div>\n</div>"
	}
	toReturn := fmt.Sprintf(start+`
<div class="sect%d">
<h%d id="%s">%s</h%d>
<div class="sectionbody">`,
		content.HeadingLevel-1,       // CSS class
		content.HeadingLevel,         // HTML open tag
		extractID(content.Heading),   // ID
		processText(content.Heading), // Actual title
		content.HeadingLevel,         // HTML close tag
	)
	e.inHeading = true
	return toReturn
}

// paragraph gives us a paragraph html representation
func (e ExporterHTML) Paragraph(content *yunyun.Content) string {
	return fmt.Sprintf(
		`
<div class="paragraph%s">
<p>
%s
</p>
</div>`,
		// div class
		func() string {
			if content.IsQuote() {
				return " quote"
			}
			if content.IsCentered() {
				return " center"
			}
			if content.IsDropCap() {
				return " dropcap"
			}
			return ""
		}(),
		processText(content.Paragraph),
	)
}

// makeListItem makes an html item
func makeListItem(s string) string {
	return fmt.Sprintf(`
<li>
<p>
%s
</p>
</li>`, processText(s))
}

// list gives us a list html representation
func (e ExporterHTML) List(content *yunyun.Content) string {
	// Hijack this type for galleries
	if yunyun.HasFlag(&content.Options, yunyun.InGalleryFlag) {
		return e.gallery(content)
	}
	return fmt.Sprintf(`
<div class="ulist">
<ul>
%s
</ul>
</div>
`, strings.Join(gana.Map(makeListItem, content.List), "\n"))
}

// makeFlexItem will make an item of the flexbox .gallery with 1/3 width
func makeFlexItem(s string, folder string) string {
	return fmt.Sprintf(`<img class="item" height="33%%" width="33%%" src="%s">`, folder+s)
}

// gallery will create a flexbox gallery as defined in .gallery css class
func (e ExporterHTML) gallery(content *yunyun.Content) string {
	makeFlexItemWithFolder := func(f string) func(string) string {
		return func(s string) string {
			return makeFlexItem(s, f)
		}
	}(func() string {
		if len(content.Summary) > 0 {
			return content.Summary + "/"
		}
		return ""
	}())
	return fmt.Sprintf(`
<center>
<div class="gallery">
%s
</div>
</center>
`, strings.Join(gana.Map(makeFlexItemWithFolder, content.List), "\n"))
}

// listNumbered gives us a numbered list html representation
func (e ExporterHTML) ListNumbered(content *yunyun.Content) string {
	// TODO
	return ""
}

// sourceCodeLang maps our source code language name to the
// name that highlight.js will need when coloring code
var sourceCodeLang = map[string]string{
	"":   "plaintext",
	"sh": "bash",
}

// mapSourceCodeLang tries to map the simple source code language
// to the one that highlight.js would accept
func mapSourceCodeLang(s string) string {
	if v, ok := sourceCodeLang[s]; ok {
		return v
	}
	return s
}

// sourceCode gives us a source code html representation
func (e ExporterHTML) SourceCode(content *yunyun.Content) string {
	lang := mapSourceCodeLang(content.SourceCodeLang)
	return fmt.Sprintf(`
<div class="listingblock">
<div class="content">
<pre class="highlight"><code class="language-%s" data-lang="%s">%s</code></pre>
</div>
</div>
`, lang, lang, func() string {
		// Remove the nested parser blockers
		s := strings.ReplaceAll(content.SourceCode, ",#", "#")
		// Escape the whatever HTML that is found in source code
		s = html.EscapeString(s)
		return s
	}())
}

// rawHTML gives us a raw html representation
func (e ExporterHTML) RawHTML(content *yunyun.Content) string {
	// If the unsafe flag is enabled, don't even wrap it in `mediablock`
	if yunyun.HasFlag(&content.Options, yunyun.InRawHtmlFlagUnsafe) {
		return content.RawHTML
	}
	return fmt.Sprintf(rawHTMLTemplate, content.RawHTML, content.Caption)

}

// horizontalLine gives us a horizontal line html representation
func (e ExporterHTML) HorizontalLine(content *yunyun.Content) string {
	return `<hr>`
}

// attentionBlock gives us a attention block html representation
func (e ExporterHTML) AttentionBlock(content *yunyun.Content) string {
	return fmt.Sprintf(`
<div class="admonitionblock note">
<table>
<tr>
<td class="icon">
<div class="title">%s</div>
</td>
<td class="content">
%s
</td>
</tr>
</table>
</div>`, content.AttentionTitle, processText(content.AttentionText))
}

// table gives an HTML formatted table
func (e ExporterHTML) Table(content *yunyun.Content) string {
	rows := make([]string, len(content.Table))
	for i := range content.Table {
		for j, v := range content.Table[i] {
			topTag := "td"
			if i == 0 && content.TableHeaders {
				topTag = "th"
			}
			content.Table[i][j] = fmt.Sprintf("<%s>%s</%s>", topTag, processText(v), topTag)
		}
		rows[i] = fmt.Sprintf("<tr>\n%s</tr>", strings.Join(content.Table[i], "\n"))
	}
	tableHTML := fmt.Sprintf("<table>%s</table>", strings.Join(rows, "\n"))
	return fmt.Sprintf(tableTemplate, content.Caption, tableHTML)
}

// table gives an HTML formatted table
func (e ExporterHTML) Details(content *yunyun.Content) string {
	if yunyun.HasFlag(&content.Options, yunyun.InDetailsFlag) {
		return fmt.Sprintf("<details>\n<summary>%s</summary>\n<hr>", content.Summary)
	}
	return "</details>"
}
