// Package github provides utilities for listing and cloning GitHub repositories
// so that go-cloc can count lines of code without requiring manual clones.
package github

import (
	"encoding/json"
	"fmt"
	"go-cloc/logger"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Repository represents the subset of GitHub REST API repository fields we
// care about.
type Repository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"` // HTTPS clone URL
	Private  bool   `json:"private"`
	Archived bool   `json:"archived"`
	Fork     bool   `json:"fork"`
}

// CloneResult holds the outcome of an individual clone operation.
type CloneResult struct {
	Repo    Repository
	ClonDir string
	Err     error
}

// ListOrgRepos fetches every repository that belongs to the given GitHub
// organisation.  It automatically pages through all results.
//
// If token is non-empty it is used as a Bearer token for authentication,
// which also reveals private repositories that the token owner can access.
func ListOrgRepos(org, token string) ([]Repository, error) {
	var allRepos []Repository
	page := 1
	const perPage = 100

	for {
		url := fmt.Sprintf(
			"https://api.github.com/orgs/%s/repos?per_page=%d&page=%d&type=all",
			org, perPage, page,
		)
		repos, err := fetchRepoList(url, token)
		if err != nil {
			return nil, fmt.Errorf("listing repos for org %q (page %d): %w", org, page, err)
		}
		allRepos = append(allRepos, repos...)
		if len(repos) < perPage {
			break
		}
		page++
	}
	return allRepos, nil
}

// GetRepo fetches metadata for a single repository identified by its full
// name, e.g. "owner/repo".
func GetRepo(fullName, token string) (Repository, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", fullName)
	req, err := newRequest(url, token)
	if err != nil {
		return Repository{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Repository{}, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Repository{}, fmt.Errorf("GitHub API returned %d for %s: %s", resp.StatusCode, url, string(body))
	}

	var repo Repository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return Repository{}, fmt.Errorf("decoding repo response: %w", err)
	}
	return repo, nil
}

// CloneRepos clones a slice of repositories into cloneDir using up to
// concurrency parallel workers.  It returns one CloneResult per repository;
// failures are recorded in CloneResult.Err rather than aborting the whole
// run so that one bad repo does not block the others.
//
// --depth=1 (shallow clone) is used to minimise network and disk usage when
// only LOC counting is needed.
//
// branch controls which branch is cloned. Pass an empty string to use each
// repository's own default branch.
func CloneRepos(repos []Repository, cloneDir, token, branch string, concurrency int) []CloneResult {
	if concurrency < 1 {
		concurrency = 1
	}

	work := make(chan Repository, len(repos))
	for _, r := range repos {
		work <- r
	}
	close(work)

	resultsChan := make(chan CloneResult, len(repos))

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range work {
				dir, err := cloneRepo(repo, cloneDir, token, branch)
				resultsChan <- CloneResult{Repo: repo, ClonDir: dir, Err: err}
			}
		}()
	}
	wg.Wait()
	close(resultsChan)

	var results []CloneResult
	for r := range resultsChan {
		results = append(results, r)
	}
	return results
}

// FilterRepos applies common filtering rules to a list of repositories.
func FilterRepos(repos []Repository, skipArchived, skipForks bool) []Repository {
	var filtered []Repository
	for _, r := range repos {
		if skipArchived && r.Archived {
			logger.Info("Skipping archived repo: ", r.FullName)
			continue
		}
		if skipForks && r.Fork {
			logger.Info("Skipping forked repo: ", r.FullName)
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

// ──────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────────────────

func fetchRepoList(url, token string) ([]Repository, error) {
	req, err := newRequest(url, token)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("decoding repo list: %w", err)
	}
	return repos, nil
}

func newRequest(url, token string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}

// cloneRepo clones a single repository with a shallow (depth=1) clone.
// If branch is empty the repository's default branch is used.
// If the destination directory already exists the clone is skipped.
func cloneRepo(repo Repository, cloneDir, token, branch string) (string, error) {
	repoDir := filepath.Join(cloneDir, repo.Name)

	if _, err := os.Stat(repoDir); err == nil {
		logger.Info("Already exists, skipping clone: ", repoDir)
		return repoDir, nil
	}

	cloneURL := repo.CloneURL
	if token != "" {
		cloneURL = embedToken(cloneURL, token)
	}

	args := []string{"clone", "--depth=1"}
	if branch != "" {
		args = append(args, "--branch", branch)
		logger.Info("Cloning ", repo.FullName, " (branch: ", branch, ") …")
	} else {
		logger.Info("Cloning ", repo.FullName, " (default branch) …")
	}
	args = append(args, cloneURL, repoDir)

	cmd := exec.Command("git", args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone %s: %w", repo.FullName, err)
	}
	logger.Info("Cloned  ", repo.FullName, " → ", repoDir)
	return repoDir, nil
}

// embedToken rewrites an HTTPS clone URL to include a personal-access-token
// so that private repositories can be cloned without an SSH key.
//
// https://github.com/owner/repo  →  https://<token>@github.com/owner/repo
func embedToken(cloneURL, token string) string {
	const prefix = "https://"
	if strings.HasPrefix(cloneURL, prefix) {
		return prefix + token + "@" + cloneURL[len(prefix):]
	}
	return cloneURL
}
