package validator

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/yaml.v3"
)

// SourceType represents the type of rule source.
type SourceType string

const (
	SourceLocal SourceType = "local"
	SourceHTTP  SourceType = "http"
	SourceGit   SourceType = "git"
)

// ParsedSource contains the parsed information from a source string.
type ParsedSource struct {
	Type     SourceType
	URL      string
	Ref      string // branch or tag for Git
	Host     string // github, gitlab, bitbucket for auth
	CacheKey string
}

// CacheMetadata stores information about a cached source.
type CacheMetadata struct {
	Source    string    `json:"source"`
	FetchedAt time.Time `json:"fetched_at"`
	TTL       string    `json:"ttl"`
}

// RuleSourceConfig represents the rules section in .promener.yaml
type RuleSourceConfig struct {
	Rules []string `yaml:"rules"`
}

// RuleSourceResolver loads Rego rules from various sources.
type RuleSourceResolver struct {
	cacheDir string
	cacheTTL time.Duration
	visited  map[string]bool
}

// gitHosts maps shorthand prefixes to Git URLs.
var gitHosts = map[string]string{
	"github":    "https://github.com/%s.git",
	"gitlab":    "https://gitlab.com/%s.git",
	"bitbucket": "https://bitbucket.org/%s.git",
}

// gitSourcePattern matches git source formats like "github:org/repo@tag" or "github:org/repo#branch"
var gitSourcePattern = regexp.MustCompile(`^(github|gitlab|bitbucket):([^@#]+)(?:[@#](.+))?$`)

// NewRuleSourceResolver creates a new resolver with default settings.
func NewRuleSourceResolver() *RuleSourceResolver {
	return &RuleSourceResolver{
		cacheDir: getCacheDir(),
		cacheTTL: time.Hour,
		visited:  make(map[string]bool),
	}
}

// getCacheDir returns the cache directory, respecting PROMENER_CACHE_DIR env var.
func getCacheDir() string {
	if dir := os.Getenv("PROMENER_CACHE_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "promener", "cache")
	}
	return filepath.Join(home, ".promener", "cache")
}

// ParseSource parses a source string and returns its components.
func ParseSource(source string) (*ParsedSource, error) {
	// Check for Git shorthand (github:, gitlab:, bitbucket:)
	if matches := gitSourcePattern.FindStringSubmatch(source); matches != nil {
		host := matches[1]
		path := matches[2]
		ref := matches[3]
		if ref == "" {
			ref = "HEAD"
		}

		urlTemplate, ok := gitHosts[host]
		if !ok {
			return nil, fmt.Errorf("unknown git host: %s", host)
		}

		return &ParsedSource{
			Type:     SourceGit,
			URL:      fmt.Sprintf(urlTemplate, path),
			Ref:      ref,
			Host:     host,
			CacheKey: hashSource(source),
		}, nil
	}

	// Check for HTTP/HTTPS URLs
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return &ParsedSource{
			Type:     SourceHTTP,
			URL:      source,
			CacheKey: hashSource(source),
		}, nil
	}

	// Default to local path
	absPath, err := filepath.Abs(source)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	return &ParsedSource{
		Type:     SourceLocal,
		URL:      absPath,
		CacheKey: hashSource(absPath),
	}, nil
}

// hashSource creates a SHA256 hash of the source for cache key.
func hashSource(source string) string {
	h := sha256.New()
	h.Write([]byte(source))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Load resolves a source and returns paths to all .rego files.
func (r *RuleSourceResolver) Load(ctx context.Context, source string) ([]string, error) {
	parsed, err := ParseSource(source)
	if err != nil {
		return nil, err
	}

	// Cycle detection
	if r.visited[parsed.CacheKey] {
		return nil, nil
	}
	r.visited[parsed.CacheKey] = true

	// Resolve source to a local directory
	dir, err := r.resolveToDir(ctx, parsed)
	if err != nil {
		return nil, err
	}

	// Check for .promener.yaml in the directory
	configPath := filepath.Join(dir, ".promener.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return r.loadFromConfig(ctx, dir, configPath)
	}

	// No config, collect .rego files directly
	return r.collectRegoFiles(dir)
}

// resolveToDir resolves a source to a local directory.
func (r *RuleSourceResolver) resolveToDir(ctx context.Context, parsed *ParsedSource) (string, error) {
	switch parsed.Type {
	case SourceLocal:
		if _, err := os.Stat(parsed.URL); os.IsNotExist(err) {
			return "", fmt.Errorf("local path does not exist: %s", parsed.URL)
		}
		return parsed.URL, nil

	case SourceHTTP:
		return r.fetchHTTP(ctx, parsed)

	case SourceGit:
		return r.cloneGit(ctx, parsed)

	default:
		return "", fmt.Errorf("unknown source type: %s", parsed.Type)
	}
}

// loadFromConfig reads .promener.yaml and resolves rules recursively.
func (r *RuleSourceResolver) loadFromConfig(ctx context.Context, baseDir, configPath string) ([]string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config RuleSourceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	var allFiles []string
	for _, rule := range config.Rules {
		// Resolve relative paths from baseDir
		if !filepath.IsAbs(rule) && !strings.Contains(rule, ":") {
			rule = filepath.Join(baseDir, rule)
		}

		files, err := r.Load(ctx, rule)
		if err != nil {
			return nil, fmt.Errorf("failed to load rule %s: %w", rule, err)
		}
		allFiles = append(allFiles, files...)
	}

	return allFiles, nil
}

// collectRegoFiles walks a directory and collects all .rego files.
func (r *RuleSourceResolver) collectRegoFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".rego" {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// fetchHTTP downloads and extracts rules from an HTTP source.
func (r *RuleSourceResolver) fetchHTTP(ctx context.Context, parsed *ParsedSource) (string, error) {
	cacheDir := filepath.Join(r.cacheDir, "http", parsed.CacheKey)

	// Check cache validity
	if r.isCacheValid(cacheDir) {
		return cacheDir, nil
	}

	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache dir: %w", err)
	}

	// Download
	req, err := http.NewRequestWithContext(ctx, "GET", parsed.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", parsed.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d for %s", resp.StatusCode, parsed.URL)
	}

	// Extract based on content type or URL
	if strings.HasSuffix(parsed.URL, ".tar.gz") || strings.HasSuffix(parsed.URL, ".tgz") {
		if err := extractTarGz(resp.Body, cacheDir); err != nil {
			return "", fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	} else if strings.HasSuffix(parsed.URL, ".zip") {
		if err := extractZip(resp.Body, cacheDir); err != nil {
			return "", fmt.Errorf("failed to extract zip: %w", err)
		}
	} else if strings.HasSuffix(parsed.URL, ".rego") {
		// Single file download
		outPath := filepath.Join(cacheDir, filepath.Base(parsed.URL))
		out, err := os.Create(outPath)
		if err != nil {
			return "", err
		}
		defer out.Close()
		if _, err := io.Copy(out, resp.Body); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("unsupported HTTP content type for %s", parsed.URL)
	}

	// Write cache metadata
	r.writeCacheMetadata(cacheDir, parsed.URL)

	return cacheDir, nil
}

// cloneGit clones a Git repository.
func (r *RuleSourceResolver) cloneGit(ctx context.Context, parsed *ParsedSource) (string, error) {
	cacheDir := filepath.Join(r.cacheDir, "git", parsed.CacheKey)

	// Check cache validity
	if r.isCacheValid(cacheDir) {
		return cacheDir, nil
	}

	// Remove old cache if exists
	os.RemoveAll(cacheDir)

	// Prepare clone options
	cloneOpts := &git.CloneOptions{
		URL:   parsed.URL,
		Depth: 1,
	}

	// Set reference if not HEAD
	if parsed.Ref != "" && parsed.Ref != "HEAD" {
		// Determine if it's a tag (@) or branch (#)
		if strings.Contains(parsed.Ref, ".") || strings.HasPrefix(parsed.Ref, "v") {
			cloneOpts.ReferenceName = plumbing.NewTagReferenceName(parsed.Ref)
		} else {
			cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(parsed.Ref)
		}
		cloneOpts.SingleBranch = true
	}

	// Set authentication if available
	if auth := getGitAuth(parsed.Host); auth != nil {
		cloneOpts.Auth = auth
	}

	// Clone
	_, err := git.PlainCloneContext(ctx, cacheDir, false, cloneOpts)
	if err != nil {
		return "", fmt.Errorf("failed to clone %s: %w", parsed.URL, err)
	}

	// Write cache metadata
	r.writeCacheMetadata(cacheDir, fmt.Sprintf("%s@%s", parsed.URL, parsed.Ref))

	return cacheDir, nil
}

// getGitAuth returns authentication for Git based on environment variables.
func getGitAuth(host string) *githttp.BasicAuth {
	// Try host-specific token first
	token := os.Getenv(strings.ToUpper(host) + "_TOKEN")
	if token == "" {
		token = os.Getenv("GIT_TOKEN")
	}
	if token == "" {
		return nil
	}

	return &githttp.BasicAuth{
		Username: "x-access-token",
		Password: token,
	}
}

// isCacheValid checks if cache exists and is not expired.
func (r *RuleSourceResolver) isCacheValid(cacheDir string) bool {
	metaPath := filepath.Join(cacheDir, "metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return false
	}

	var meta CacheMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return false
	}

	return time.Since(meta.FetchedAt) < r.cacheTTL
}

// writeCacheMetadata writes cache metadata file.
func (r *RuleSourceResolver) writeCacheMetadata(cacheDir, source string) {
	meta := CacheMetadata{
		Source:    source,
		FetchedAt: time.Now(),
		TTL:       r.cacheTTL.String(),
	}

	data, _ := json.MarshalIndent(meta, "", "  ")
	os.WriteFile(filepath.Join(cacheDir, "metadata.json"), data, 0644)
}

// extractTarGz extracts a tar.gz archive to the destination directory.
func extractTarGz(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		// Security: prevent path traversal
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid tar path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

// extractZip extracts a zip archive to the destination directory.
func extractZip(r io.Reader, dest string) error {
	// Read all content to a temp file (zip needs random access)
	tmpFile, err := os.CreateTemp("", "promener-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, r); err != nil {
		return err
	}

	// Open as zip
	zipReader, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		target := filepath.Join(dest, f.Name)

		// Security: prevent path traversal
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid zip path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
