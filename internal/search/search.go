// Copyright 2022-2026 Salesforce, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package search

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const SiteURL = "https://docs.slack.dev"

// SearchResult represents a single search result
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Type    string `json:"type"`
	Score   int    `json:"-"` // Used for sorting, not exported to JSON
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Query   string         `json:"query"`
	Filter  string         `json:"filter"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Showing int            `json:"showing"`
}

// FrontMatter represents the YAML frontmatter in markdown files
type FrontMatter struct {
	Title    string
	Unlisted bool
}

// parseFrontMatter extracts frontmatter from markdown content
func parseFrontMatter(content string) (*FrontMatter, string) {
	// Check if content starts with frontmatter
	if !strings.HasPrefix(content, "---\n") {
		return &FrontMatter{}, content
	}

	// Find the closing ---
	lines := strings.Split(content, "\n")
	var endIndex int
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIndex = i
			break
		}
	}

	if endIndex == 0 {
		return &FrontMatter{}, content
	}

	// Parse frontmatter lines
	fm := &FrontMatter{}
	for i := 1; i < endIndex; i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "title:") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			// Remove quotes if present
			title = strings.Trim(title, `"'`)
			fm.Title = title
		} else if strings.HasPrefix(line, "unlisted:") {
			unlisted := strings.TrimSpace(strings.TrimPrefix(line, "unlisted:"))
			fm.Unlisted = unlisted == "true"
		}
	}

	// Return body content (everything after frontmatter)
	bodyLines := lines[endIndex+1:]
	body := strings.Join(bodyLines, "\n")
	return fm, body
}

// extractTitle attempts to extract title from markdown content
func extractTitle(content string) string {
	// Try H1 heading
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}

	// Try HTML h1
	re := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return stripHTML(matches[1])
	}

	return ""
}

// stripHTML removes HTML tags and markdown formatting
func stripHTML(text string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text = re.ReplaceAllString(text, "")

	// Replace HTML entities
	replacements := map[string]string{
		"&nbsp;": " ",
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
	}

	for entity, replacement := range replacements {
		text = strings.ReplaceAll(text, entity, replacement)
	}

	// Remove markdown links [text](url)
	re = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	text = re.ReplaceAllString(text, "$1")

	// Remove inline code `code`
	re = regexp.MustCompile("`([^`]+)`")
	text = re.ReplaceAllString(text, "$1")

	// Remove bold/italic *text* and **text**
	re = regexp.MustCompile(`\*{1,2}([^*]+)\*{1,2}`)
	text = re.ReplaceAllString(text, "$1")

	// Normalize whitespace
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// extractSnippet finds text around the query term
func extractSnippet(content, query string, maxLength int) string {
	cleanContent := stripHTML(content)
	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(cleanContent)

	queryIndex := strings.Index(contentLower, queryLower)
	if queryIndex == -1 {
		// No match, return beginning
		if len(cleanContent) > maxLength {
			return cleanContent[:maxLength] + "..."
		}
		return cleanContent
	}

	// Extract context around the match
	start := queryIndex - 100
	if start < 0 {
		start = 0
	}

	end := queryIndex + len(query) + 150
	if end > len(cleanContent) {
		end = len(cleanContent)
	}

	snippet := cleanContent[start:end]

	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(cleanContent) {
		snippet = snippet + "..."
	}

	return strings.TrimSpace(snippet)
}

// calculateRelevance scores a document based on query matches
func calculateRelevance(content, title, query string) int {
	queryLower := strings.ToLower(query)
	titleLower := strings.ToLower(title)
	contentLower := strings.ToLower(content)

	score := 0

	// Title matches are highly relevant
	if strings.Contains(titleLower, queryLower) {
		score += 100
		if titleLower == queryLower {
			score += 50 // Exact title match
		}
	}

	// Count occurrences in content
	matches := strings.Count(contentLower, queryLower)
	score += matches * 10

	// Boost for early occurrence
	firstIndex := strings.Index(contentLower, queryLower)
	if firstIndex != -1 && firstIndex < 500 {
		score += 20
	}

	return score
}

// filePathToURL converts a file path to a docs URL
func filePathToURL(filePath, contentDir string) string {
	relPath, err := filepath.Rel(contentDir, filePath)
	if err != nil {
		return "/"
	}

	// Remove .md extension
	relPath = strings.TrimSuffix(relPath, ".md")

	// Handle index files
	if strings.HasSuffix(relPath, "/index") {
		relPath = strings.TrimSuffix(relPath, "/index")
	} else if relPath == "index" {
		return "/"
	}

	// Convert to URL path
	urlPath := "/" + strings.ReplaceAll(relPath, "\\", "/")
	return urlPath
}

// determineType determines content type from file path
func determineType(filePath string) string {
	if strings.Contains(filePath, "/reference/") {
		return "reference"
	}
	if strings.Contains(filePath, "/changelog/") {
		return "changelog"
	}
	if strings.Contains(filePath, "/tools/") {
		return "tools"
	}
	if strings.Contains(filePath, "/apis/") {
		return "api"
	}
	return "guide"
}

// matchesFilter checks if a file matches the given filter
func matchesFilter(filePath, filter string) bool {
	if filter == "" || filter == "all" {
		return true
	}

	contentType := determineType(filePath)

	switch filter {
	case "reference":
		return contentType == "reference"
	case "guides", "guide":
		return contentType == "guide"
	case "changelog":
		return contentType == "changelog"
	case "tools":
		return contentType == "tools"
	case "apis", "api":
		return contentType == "api"
	default:
		return true
	}
}

// findMarkdownFiles recursively finds all .md files in a directory
func findMarkdownFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// SearchDocs performs a programmatic search of documentation files
func SearchDocs(query, filter string, limit int, contentDir string) (*SearchResponse, error) {
	if contentDir == "" {
		return nil, fmt.Errorf("content directory not specified")
	}

	// Check if content directory exists
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return &SearchResponse{
			Query:   query,
			Filter:  filter,
			Results: []SearchResult{},
			Total:   0,
			Showing: 0,
		}, fmt.Errorf("content directory not found: %s", contentDir)
	}

	// Find all markdown files
	markdownFiles, err := findMarkdownFiles(contentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find markdown files: %w", err)
	}

	var results []SearchResult
	queryLower := strings.ToLower(query)

	// Search through files
	for _, filePath := range markdownFiles {
		// Apply filter
		if !matchesFilter(filePath, filter) {
			continue
		}

		// Read file
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files we can't read
		}

		contentStr := string(content)

		// Parse frontmatter
		frontmatter, bodyContent := parseFrontMatter(contentStr)

		// Skip unlisted pages
		if frontmatter.Unlisted {
			continue
		}

		// Check if query matches (case insensitive)
		if !strings.Contains(strings.ToLower(contentStr), queryLower) {
			continue
		}

		// Extract metadata
		title := frontmatter.Title
		if title == "" {
			title = extractTitle(bodyContent)
		}
		if title == "" {
			title = "Untitled"
		}

		url := SiteURL + filePathToURL(filePath, contentDir)
		contentType := determineType(filePath)
		snippet := extractSnippet(bodyContent, query, 250)
		score := calculateRelevance(contentStr, title, query)

		result := SearchResult{
			Title:   title,
			URL:     url,
			Snippet: snippet,
			Type:    contentType,
			Score:   score,
		}

		results = append(results, result)
	}

	// Sort by relevance score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	total := len(results)
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	if filter == "" {
		filter = "all"
	}

	response := &SearchResponse{
		Query:   query,
		Filter:  filter,
		Results: results,
		Total:   total,
		Showing: len(results),
	}

	return response, nil
}

// FindDocsRepo attempts to locate the docs repository
func FindDocsRepo() string {
	candidates := []string{
		"../docs",
		"../../docs",
		"./docs",
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}

		contentDir := filepath.Join(absPath, "content")
		if _, err := os.Stat(contentDir); err == nil {
			return absPath
		}
	}

	return ""
}
