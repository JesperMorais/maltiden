package domain

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrContainsHTML      = errors.New("contains_html")
	ErrContainsURL       = errors.New("contains_url")
	ErrContainsMarkdown  = errors.New("contains_markdown")
	ErrContainsScraping  = errors.New("contains_scraping")
	ErrContainsProfanity = errors.New("contains_profanity")
	ErrContainsInjection = errors.New("contains_injection")
)

var (
	reHTMLEntity = regexp.MustCompile(`&amp;|&nbsp;|&quot;|&#\d+`)
	reHTMLTag    = regexp.MustCompile(`<[^>]+>`)

	reURL = regexp.MustCompile(`(?i)https?://|www\.`)

	reMarkdownLink   = regexp.MustCompile(`!?\[[^\]]*\]\([^)]*\)`)
	reMarkdownEmph   = regexp.MustCompile(`\*\*[^*]+\*\*|__[^_]+__` + "|`[^`]+`")
	reMarkdownHeader = regexp.MustCompile(`(?m)^[ \t]*[#>]\s`)

	reScraping = regexp.MustCompile(`(?i)klicka här|läs mer|click here|read more`)

	reProfanity = regexp.MustCompile(`(?i)\b(bajs|skit|kuk|fitt|helvete)\b`)

	reInjection = regexp.MustCompile(
		`(?im)\bIGNORE (?:PRIOR|PREVIOUS)\b` +
			`|\bDISREGARD (?:PRIOR|PREVIOUS)\b` +
			`|\bOUTPUT ONLY\b` +
			`|\bRESPOND ONLY\b` +
			`|</system>` +
			`|<\|im_(?:start|end)\|>` +
			`|\[INST\]` +
			`|\[/INST\]` +
			`|^(?:SYSTEM|ASSISTANT|USER):` +
			`|^(?:act as|pretend you are|you are now)`,
	)
)

// ValidateContent returns the first content policy violation found in s, or nil if clean.
// Injection is checked first because some injection tokens (</system>, <|im_*|>) also
// match the HTML pattern — we want the more specific error.
// Markdown is checked before URL because markdown links embed URLs inside them.
func ValidateContent(s string) error {
	if reInjection.MatchString(s) {
		return ErrContainsInjection
	}
	if reHTMLEntity.MatchString(s) || reHTMLTag.MatchString(s) {
		return ErrContainsHTML
	}
	if reMarkdownLink.MatchString(s) || reMarkdownEmph.MatchString(s) || reMarkdownHeader.MatchString(s) {
		return ErrContainsMarkdown
	}
	if reURL.MatchString(s) {
		return ErrContainsURL
	}
	if reScraping.MatchString(s) {
		return ErrContainsScraping
	}
	if reProfanity.MatchString(s) {
		return ErrContainsProfanity
	}
	return nil
}

// StripEmoji removes emoji runes (So, Sk categories; ZWJ; variation selector;
// regional indicators) while preserving letters, marks, numbers, and punctuation.
func StripEmoji(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isEmojiRune(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isEmojiRune(r rune) bool {
	switch {
	case r == 0x200D: // ZWJ
		return true
	case r == 0xFE0F: // variation selector-16
		return true
	case r >= 0x1F1E6 && r <= 0x1F1FF: // regional indicators
		return true
	default:
		return unicode.Is(unicode.So, r) || unicode.Is(unicode.Sk, r)
	}
}
