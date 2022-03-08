package orgmode

import (
	"darkness/internals"
	"strings"
)

func Parse(lines []string) *internals.Page {
	page := &internals.Page{}
	page.Contents = make([]internals.Content, 0, 16)

	currentContext := ""
	inList := false
	currentList := make([]string, 0, 8)
	for i, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if isComment(line) {
			continue
		}
		previousContext := currentContext
		currentContext = currentContext + line
		// If it's an empty line, then process current text
		if line == "" {
			// Empty context
			if previousContext == "" {
				continue
			}
			// Let's see if our context is a standalone link
			if link := isLink(currentContext); link != nil {
				page.Contents = append(page.Contents, *link)
				currentContext = ""
				continue
			}
			// New line break means we have to save the paragraph
			// we just read if we're not currently reading a list
			if !inList {
				page.Contents = append(
					page.Contents,
					*formParagraph(strings.TrimSpace(currentContext)))
			}
			currentContext = ""
		}
		// We are in a list now and it's not the first line (reserved for title)
		if isList(line) && i != 0 {
			// If we were not in a list context before, save what we have
			if !inList && len(previousContext) > 0 {
				page.Contents = append(
					page.Contents,
					*formParagraph(strings.TrimSpace(previousContext)))
				currentContext = ""
			}
			// Mark that we entered a list context
			inList = true
			// Trim the bullet points with [2:]
			currentList = append(currentList, line[2:])
		} else if inList {
			// We are not in a list anymore right now but we were right
			// before this, it means we have to save the list we just read
			page.Contents = append(page.Contents, internals.Content{
				Type: internals.TypeList,
				List: currentList,
			})
			// Empty the tracker
			currentList = []string{}
			// Mark that we left the list context
			inList = false
			// Restore the context
			currentContext = ""
		}
		// Check if we are in a source block
		// Find whether the current line is a part of a list
		// A header is found, append and continue
		if header := isHeader(line); header != nil &&
			(((i == 0) && header.HeaderLevel == 1) || header.HeaderLevel > 1) {
			currentContext = ""
			// Level 1 is the page title
			if header.HeaderLevel == 1 {
				page.Title = header.Header
				continue
			}
			page.Contents = append(page.Contents, *header)
			continue
		}
		currentContext += " "
	}
	return page
}

func isHeader(line string) *internals.Content {
	level := 0
	for _, c := range line {
		if c != '*' {
			break
		}
		level++
	}
	// Not a header
	if level < 1 {
		return nil
	}
	// Is a header
	return &internals.Content{
		Type:        internals.TypeHeader,
		HeaderLevel: level,
		Header:      line[level+1:],
	}
}

func isComment(line string) bool {
	return strings.HasPrefix(line, "# ")
}

func isLink(line string) *internals.Content {
	line = strings.TrimSpace(line)
	// Not a link
	if !LinkRegexp.MatchString(line) {
		return nil
	}
	submatches := LinkRegexp.FindAllStringSubmatch(line, 1)
	// Sanity check
	if len(submatches) < 1 {
		return nil
	}
	match := strings.TrimSpace(submatches[0][0])
	link := strings.TrimSpace(submatches[0][1])
	text := strings.TrimSpace(submatches[0][2])
	// Check if this is a standalone link (just by itself on a line)
	// If it's not, then it's a simple link in a paragraph, deal with
	// it later in `htmlize`
	if len(match) != len(line) {
		return nil
	}
	content := &internals.Content{
		Type:      internals.TypeLink,
		Link:      link,
		LinkTitle: text,
	}
	// Our link is standalone. Check if it's an image
	if strings.HasSuffix(link, ".png") {
		content.Type = internals.TypeImage
		content.ImageSource = link
		content.ImageCaption = text
		return content
	}
	// Check if it's a youtube video embed
	if strings.HasPrefix(link, "https://youtu.be/") {
		content.Type = internals.TypeYoutube
		content.Youtube = link[17:]
		return content
	}
	// Check if it's a spotify track link
	if strings.HasPrefix(link, "https://open.spotify.com/track/") {
		content.Type = internals.TypeSpotifyTrack
		content.SpotifyTrack = link[31:]
		return content
	}
	// Check if it's a spotify playlist link
	if strings.HasPrefix(link, "https://open.spotify.com/playlist/") {
		content.Type = internals.TypeSpotifyPlaylist
		content.SpotifyPlaylist = link[34:]
		return content
	}
	return nil
}

func formParagraph(text string) *internals.Content {
	return &internals.Content{
		Type:      internals.TypeParagraph,
		Paragraph: text,
	}
}

func isList(line string) bool {
	for _, prefix := range listPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func isSourceCode(line string) bool {
	return strings.HasPrefix(line, `#+begin_src`)
}
