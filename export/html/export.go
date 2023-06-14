package html

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/thecsw/darkness/emilia"
	"github.com/thecsw/darkness/emilia/puck"
	"github.com/thecsw/darkness/yunyun"
	"github.com/thecsw/gana"
)

const (
	tombEnding = " ◼"
)

var (
	//go:embed banner.txt
	darknessBannerSource string
	// darknessBanner wrapes `darknessBannerSource` in a comment block.
	darknessBanner = "<!--\n" + darknessBannerSource + "\n-->\n"
)

// Export runs the process of exporting
func (e *ExporterHTML) Export() string {
	defer puck.Stopwatch("Exported", "page", e.page.File).Record()
	if e.page == nil {
		fmt.Println("Export should be called after SetPage")
		os.Exit(1)
	}

	// Initialize the html mapping after yunyun built regexes.
	markupHTMLMapping = map[*regexp.Regexp]string{
		yunyun.ItalicText:        `$l<em>$text</em>$r`,
		yunyun.BoldText:          `$l<strong>$text</strong>$r`,
		yunyun.VerbatimText:      `$l<code>$text</code>$r`,
		yunyun.StrikethroughText: `$l<s>$text</s>$r`,
		yunyun.UnderlineText:     `$l<u>$text</u>$r`,
		yunyun.SuperscriptText:   `$l<sup>$text</sup>$r`,
		yunyun.SubscriptText:     `$l<sub>$text</sub>$r`,
	}

	// Add the red tomb to the last paragraph on given directories.
	// Only trigger if the tombs were manually flipped.
	if e.page.Accoutrement.Tomb.IsEnabled() {
		e.addTomb()
	}
	// If the page hasn't set a custom preview, default to emilia.
	if len(e.page.Accoutrement.Preview) < 1 {
		e.page.Accoutrement.Preview = string(emilia.Config.Website.Preview)
	}

	if e.page.Accoutrement.Toc.IsEnabled() {
		e.page.Contents = append(e.toc(), e.page.Contents...)
	}

	// Build the HTML (string) representation of each content
	content := make([]string, 0, len(e.page.Contents))
	for i, v := range e.page.Contents {
		e.currentContentIndex = i
		e.currentContent = v
		content = append(content, e.buildContent(v))
	}

	return fmt.Sprintf(`%s<!DOCTYPE html>
<html lang="en">
<head>
%s
<title>%s</title>
</head>
<body class="article">
%s
%s
%s
</body>
</html>`,
		darknessBanner,
		e.combineAndFilterHtmlHead(),
		processTitle(flattenFormatting(e.page.Title)),
		e.authorHeader(),
		strings.Join(content, ""),
		e.addFootnotes(),
	)
}

// buildContent builds the HTML representation of a content.
func (e *ExporterHTML) buildContent(content *yunyun.Content) string {
	// Build the HTML (string) representation of each content.
	built := e.contentFunctions[e.currentContent.Type](e.currentContent)

	// Set the content flags, like whether it's in writing mode or not.
	e.setContentFlags(e.currentContent)

	// If the content is in writing mode, wrap it in a writing div.
	// otherwise, wrap it in other divs, depending on the content type.
	return e.resolveDivTags(built)
}

// leftHeading leaves the heading.
func (e *ExporterHTML) leftHeading() {
	e.inHeading = false
}

func (e ExporterHTML) combineAndFilterHtmlHead() string {
	// Build the array of all head elements (except page's specific head options).
	allHead := [][]string{e.linkTags(), e.metaTags(), e.styleTags(), e.scriptTags(), emilia.Config.Website.ExtraHead}
	// Go through all the head elements and filter them out depending on page's specific exclusion rules.
	finalHead := ""
	for _, head := range allHead {
		finalHead += strings.Join(gana.Filter(e.page.Accoutrement.ExcludeHtmlHeadContains.ShouldKeep, head), "\n")
	}
	// Page's specific html head elements are not filtered out.
	return finalHead + strings.Join(e.page.HtmlHead, "\n")
}

// styleTags is the processed style tags.
func (e ExporterHTML) styleTags() []string {
	content := make([]string, len(emilia.Config.Website.Styles)+len(e.page.Stylesheets))
	for i, style := range emilia.Config.Website.Styles {
		stylePath := yunyun.FullPathFile(style)
		if !strings.HasPrefix(string(style), "http") {
			stylePath = emilia.JoinPath(style)
		}
		content[i] = fmt.Sprintf(
			`<link rel="stylesheet" type="text/css" href="%s">`+"\n", stylePath,
		)
	}
	return append(content, e.page.Stylesheets...)
}

// defaultScripts are the default scripts.
var defaultScripts = []string{
	`<script type="module">document.documentElement.classList.remove("no-js");document.documentElement.classList.add("js");</script>`,
	`<script async src="https://sandyuraz.com/scripts/time.js"></script>`,
}

// scriptTags returns the script tags.
func (e ExporterHTML) scriptTags() []string {
	return append(defaultScripts, e.page.Scripts...)
}

func rssLink() string {
	if !emilia.Config.RSS.Enable {
		return ""
	}
	return `<span><a href="/feed.xml" class="rss-link"><img src="/assets/rss.svg" class="rss-icon"></a></span><br>` + "\n"
}

func authorName() string {
	if !emilia.Config.Author.NameEnable {
		return ""
	}
	return `<span id="author" class="author">` + emilia.Config.Author.Name + `</span><br>` + "\n"
}

func authorEmail() string {
	if !emilia.Config.Author.EmailEnable {
		return ""
	}
	return `<span id="email" class="email">` + emilia.Config.Author.Email + `</span><br>` + "\n"
}

// authorHeader returns the author header.
func (e ExporterHTML) authorHeader() string {
	content := fmt.Sprintf(`
<div class="header">
<h1 class="section-1">%s%s</h1>
<div class="menu">
%s%s%s`,
		authorImage(e.page.Accoutrement.AuthorImage), processTitle(e.page.Title),
		rssLink(), authorName(), authorEmail(),
	)
	content += `<span id="revdate">` + "\n"

	// Build the navigation links.
	navLinks := make([]string, 0, len(emilia.Config.Navigation))

	// Go through elements.
	for i := 1; i <= len(emilia.Config.Navigation); i++ {
		// Get the navigation element read from Darkness' toml.
		v := emilia.Config.Navigation[strconv.FormatInt(int64(i), 10)]
		// If the nav element wants to hide in this location, then skip it.
		if e.page.Location == v.Hide {
			continue
		}
		// Build each of the navlinks and concat the hrefs.
		navLinks = append(navLinks, fmt.Sprintf(`<a href="%s">%s</a>`,
			emilia.JoinPathGeneric[yunyun.RelativePathDir, yunyun.FullPathDir](v.Link),
			v.Title))
	}

	// Close the navigation links span.
	content += strings.Join(navLinks, " | ") + `</span>`

	// Add the Holoscene time element.
	content += `
</div>
<div id="hetime" class="menu"></div>
</div>`
	// Return the website header.
	return content
}

// authorHeader returns img element if author header image is given.
func authorImage(authorImageFlag yunyun.AccoutrementFlip) string {
	// Return nothing if it's not provided.
	if emilia.Config.Author.Image == "" || authorImageFlag.IsDisabled() {
		return ""
	}
	return fmt.Sprintf(`<img id="myface" src="%s" alt="avatar">`,
		emilia.Config.Author.ImagePreComputed)
}

// addTomb adds the tomb to the last paragraph.
func (e ExporterHTML) addTomb() {
	// Empty???
	if len(e.page.Contents) < 1 {
		return
	}
	// Find the last paragrapd and attached the tomb.
	for i := len(e.page.Contents) - 1; i >= 0; i-- {
		// Skip if it's not a paragraph.
		if !e.page.Contents[i].IsParagraph() {
			continue
		}
		// Add the tomb and break out.
		e.page.Contents[i].Paragraph += tombEnding
		break
	}
}

// toc returns the table of contents.
func (e ExporterHTML) toc() []*yunyun.Content {
	return []*yunyun.Content{
		// First, add the table of contents header.
		{
			Type:                 yunyun.TypeHeading,
			Heading:              "Table of Contents",
			HeadingLevel:         3,
			HeadingLevelAdjusted: 1,
		},
		// Then, add the table of contents.
		{
			Type: yunyun.TypeList,
			// overload the summary field to indicate
			// that this is the table of contents.
			Summary: "toc",
			List:    emilia.GenerateTableOfContents(e.page),
		},
		// Finally, add the horizontal line.
		{
			Type: yunyun.TypeHorizontalLine,
		},
	}
}
