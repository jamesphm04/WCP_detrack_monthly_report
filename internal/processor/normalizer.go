package processor

import (
	"regexp"
	"strings"
)

// RunNumberNormalizer handles normalization of run numbers
type RunNumberNormalizer struct {
	aliases map[string]string
}

// NewRunNumberNormalizer creates a new normalizer with predefined aliases
func NewRunNumberNormalizer() *RunNumberNormalizer {
	return &RunNumberNormalizer{
		aliases: map[string]string {
			// Time variations
			"8:00AM":  "8:00AM",
			"8AM":     "8:00AM",
			"10:30AM": "10:30AM",
			"12PM":    "12:00PM",
			"12:00PM": "12:00PM",
			"1PM":     "1:00PM",
			"1:00PM":  "1:00PM",

			// Route variations
			"NORTH": "NORTH",
			"SOUTH": "SOUTH",
			"GC":    "GC",
		},
	}
}

// Normalize removes dates and standardizes the format 
func (n *RunNumberNormalizer) Normalize(runNumber string) string {
	// Remove date pattern (DD/DD/DD)
	pattern := regexp.MustCompile(`\d{2}/\d{2}/\d{2}\s*`)
	cleaned := pattern.ReplaceAllString(runNumber, "")

	// Trim white space
	cleaned = strings.TrimSpace(cleaned)

	if strings.HasPrefix(cleaned, "WCP") {
		// Handle WCP prefix formats
		return n.normalizeWCPPrefix(cleaned)
	} else {
		// Handle "ROUTE TIME" prefix formats
		return n.normalizeRouteTimePrefix(cleaned)
	}
}

// extractRoute extracts the route from a string
func (n *RunNumberNormalizer) extractRoute(s string) string {
	pattern := regexp.MustCompile(`\b(NORTH|SOUTH|GC)\b`)
	match := pattern.FindString(strings.ToUpper(s))
	return match
}

// extractTime extracts the time from a string
func (n *RunNumberNormalizer) extractTime(s string) string {
	pattern := regexp.MustCompile(`\d{1,2}(?::\d{2})?\s*(?:AM|PM)`)
	match := pattern.FindString(strings.ToUpper(s))
	return strings.TrimSpace(match)
}

// normalizeRoute normalizes route names
func (n *RunNumberNormalizer) normalizeRoute(route string) string {
	route = strings.ToUpper(strings.TrimSpace(route))
	if normalized, ok := n.aliases[route]; ok {
		return normalized
	}
	return route
}

func (n *RunNumberNormalizer) normalizeTime(time string) string {
	// Remove spaces between time and AM/PM
	time = strings.ReplaceAll(strings.ToUpper(time), " ", "")

	// Check if already in normalized format
	if normalized, ok := n.aliases[time]; ok {
		return normalized
	}

	// Handle formats like "8AM" -> "8:00AM"
	pattern := regexp.MustCompile(`^(\d{1,2})(AM|PM)$`)
	if match := pattern.FindStringSubmatch(time); match != nil {
		hour := match[1]
		period := match[2]
		return hour + ":00" + period
	}

	return time
}

// normalizeWCPPrefix handles formats like "WCPNORTH - 8:00AM"
func (n *RunNumberNormalizer) normalizeWCPPrefix(s string) string {
	// Remove "WCP" prefix to decouple WCPGC
	s = strings.TrimPrefix(s, "WCP")

	// Remove hyphens and extract spaces then re-split
	s = strings.ReplaceAll(s, "-", " ")

	// Extract route and time using regex
	route := n.extractRoute(s)
	time := n.extractTime(s)

	if route == "" || time == "" {
		return "WCP" + s
	} else {
		route = n.normalizeRoute(route)
		time = n.normalizeTime(time)
		return "WCP" + route + " - " + time
	}
}

// normalizeRouteTimePrefix handles formats like "NORTH 8:00AM" or "8AM SOUTH"
func (n *RunNumberNormalizer) normalizeRouteTimePrefix(s string) string {
	route := n.extractRoute(s)
	time := n.extractTime(s)

	if route == "" || time == "" {
		return s
	}

	route = n.normalizeRoute(route)
	time = n.normalizeTime(time)

	return "WCP" + route + " - " + time
}