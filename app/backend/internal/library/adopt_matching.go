package library

import (
	"regexp"
	"strings"
)

var folderBrackets = regexp.MustCompile(`\[[^\[\]]*\]`)
var folderBracketYear = regexp.MustCompile(`^\[\s*(?:18|19|20)\d{2}\s*\]$`)

// matchingFolderName removes release annotations from lookup text, leaving
// disk paths untouched. A bracketed year remains available to the parser.
func matchingFolderName(s string) string {
	s = strings.Join(strings.Fields(strings.ReplaceAll(s, ".", " ")), " ")
	clean := folderBrackets.ReplaceAllStringFunc(s, func(part string) string {
		if folderBracketYear.MatchString(part) {
			return "[" + strings.TrimSpace(strings.Trim(part, "[]")) + "]"
		}
		return " "
	})
	clean = strings.Join(strings.Fields(clean), " ")
	title := clean
	if m := ReFolderYear.FindStringSubmatch(clean); m != nil {
		title = m[1]
	}
	// In names such as [REC] (2007), the bracketed text is the title itself.
	if strings.Trim(title, " -_()[]") == "" {
		for _, part := range folderBrackets.FindAllString(s, -1) {
			if !folderBracketYear.MatchString(part) {
				return strings.TrimSpace(part + " " + clean)
			}
		}
		return s
	}
	return clean
}
