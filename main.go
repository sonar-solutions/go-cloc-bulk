package main

import (
	"fmt"
	"go-cloc/github"
	"go-cloc/logger"
	"go-cloc/report"
	"go-cloc/scanner"
	"go-cloc/utilities"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// parse CLI arguments and store them in a struct
	args := utilities.ParseArgsFromCLI()

	isGithubMode := args.GithubOrg != "" || len(args.GithubRepos) > 0

	if isGithubMode {
		runGithubMode(args)
	} else {
		runLocalMode(args)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Local scan (original behaviour)
// ──────────────────────────────────────────────────────────────────────────────

func runLocalMode(args utilities.CLIArgs) {
	logger.Info("Scanning ", args.LocalScanFilePath, "...")
	fileScanResultsArr := scanDirectory(args.LocalScanFilePath, args.IgnorePatterns)

	fileScanResultsArr = report.SortFileScanResults(fileScanResultsArr)
	repoTotalResult := report.CalculateTotalLineOfCode(fileScanResultsArr)
	repoTotalResultSupportedOnly := report.CalculateTotalLineOfCodeSupportedOnly(fileScanResultsArr)

	writeCSVIfRequested(args.CsvFilePath, fileScanResultsArr, repoTotalResult)
	writeHTMLIfRequested(args.HtmlReportsDirectoryPath, args.LocalScanFilePath, fileScanResultsArr)

	printDisclaimer()
	logger.Info("Printing total LOC for ALL languages ...")
	report.PrintResultsToCommandLine(repoTotalResult.CodeLineCount, repoTotalResult.CommentsLineCount, repoTotalResult.BlankLineCount)
	logger.Info("Printing LOC for SUPPORTED languages ...")
	report.PrintResultsToCommandLine(repoTotalResultSupportedOnly.CodeLineCount, repoTotalResultSupportedOnly.CommentsLineCount, repoTotalResultSupportedOnly.BlankLineCount)
	printFooter(args.LocalScanFilePath, repoTotalResultSupportedOnly.CodeLineCount)

	fmt.Println(repoTotalResultSupportedOnly.CodeLineCount)
}

// ──────────────────────────────────────────────────────────────────────────────
// GitHub auto-clone mode
// ──────────────────────────────────────────────────────────────────────────────

func runGithubMode(args utilities.CLIArgs) {
	// ── Resolve clone directory ───────────────────────────────────────────────
	cloneDir := args.CloneDir
	usingTempDir := false
	if cloneDir == "" {
		var err error
		cloneDir, err = os.MkdirTemp("", "go-cloc-*")
		if err != nil {
			logger.Error("Failed to create temporary clone directory: ", err)
			os.Exit(1)
		}
		usingTempDir = true
		logger.Info("Using temporary clone directory: ", cloneDir)
	} else {
		if err := utilities.CreateDirectoryIfNotExists(cloneDir); err != nil {
			logger.Error("Failed to create clone directory: ", err)
			os.Exit(1)
		}
	}

	// ── Cleanup on exit ───────────────────────────────────────────────────────
	// When no-cleanup is false (the default), always remove cloned repos after
	// the scan — whether the clone dir was auto-created or user-supplied.
	// Set no-cleanup:true (or --no-cleanup) to keep repos on disk so that
	// re-runs skip the clone step entirely.
	if !args.NoCleanup {
		defer func() {
			logger.Info("Cleaning up cloned repositories in ", cloneDir)
			if usingTempDir {
				// Remove the whole temp dir
				os.RemoveAll(cloneDir)
			} else {
				// User supplied the dir — only remove the individual repo
				// subdirectories, leaving the parent directory intact.
				entries, err := os.ReadDir(cloneDir)
				if err == nil {
					for _, e := range entries {
						if e.IsDir() {
							os.RemoveAll(filepath.Join(cloneDir, e.Name()))
						}
					}
				}
			}
		}()
	} else {
		logger.Info("no-cleanup is set — cloned repositories will remain in ", cloneDir)
	}

	// ── Collect repositories ──────────────────────────────────────────────────
	var repos []github.Repository

	if args.GithubOrg != "" {
		logger.Info("Fetching repository list for organisation: ", args.GithubOrg)
		orgRepos, err := github.ListOrgRepos(args.GithubOrg, args.GithubToken)
		if err != nil {
			logger.Error("Failed to list org repos: ", err)
			os.Exit(1)
		}
		logger.Info("Found ", len(orgRepos), " repositories in ", args.GithubOrg)
		repos = append(repos, orgRepos...)
	}

	// Add explicitly specified repos, deduplicating against org repos already collected
	seen := make(map[string]bool)
	for _, r := range repos {
		seen[r.FullName] = true
	}
	for _, fullName := range args.GithubRepos {
		if seen[fullName] {
			continue
		}
		repo, err := github.GetRepo(fullName, args.GithubToken)
		if err != nil {
			logger.Error("Failed to fetch repo metadata for ", fullName, ": ", err)
			os.Exit(1)
		}
		repos = append(repos, repo)
		seen[fullName] = true
	}

	// ── Filter repositories ───────────────────────────────────────────────────
	repos = github.FilterRepos(repos, args.SkipArchived, args.SkipForks)
	if len(repos) == 0 {
		logger.Error("No repositories to scan after applying filters.")
		os.Exit(1)
	}
	logger.Info("Scanning ", len(repos), " repositories ...")

	// ── Clone repositories concurrently ──────────────────────────────────────
	cloneResults := github.CloneRepos(repos, cloneDir, args.GithubToken, args.Branch, args.Concurrency)

	// ── Scan each repo and accumulate results ─────────────────────────────────
	type repoScanSummary struct {
		FullName              string
		CodeLineCount         int
		CommentsLineCount     int
		BlankLineCount        int
		SupportedCodeLines    int
		AllFileScanResults    []scanner.FileScanResults
	}

	var summaries []repoScanSummary
	var allFileResults []scanner.FileScanResults

	for _, cr := range cloneResults {
		if cr.Err != nil {
			logger.Error("Skipping ", cr.Repo.FullName, " — clone error: ", cr.Err)
			continue
		}

		logger.Info("Scanning ", cr.Repo.FullName, " ...")
		fileScanResultsArr := scanDirectory(cr.ClonDir, args.IgnorePatterns)
		fileScanResultsArr = report.SortFileScanResults(fileScanResultsArr)

		total := report.CalculateTotalLineOfCode(fileScanResultsArr)
		totalSupported := report.CalculateTotalLineOfCodeSupportedOnly(fileScanResultsArr)

		summaries = append(summaries, repoScanSummary{
			FullName:           cr.Repo.FullName,
			CodeLineCount:      total.CodeLineCount,
			CommentsLineCount:  total.CommentsLineCount,
			BlankLineCount:     total.BlankLineCount,
			SupportedCodeLines: totalSupported.CodeLineCount,
			AllFileScanResults: fileScanResultsArr,
		})
		allFileResults = append(allFileResults, fileScanResultsArr...)
	}

	if len(summaries) == 0 {
		logger.Error("No repositories were successfully scanned.")
		os.Exit(1)
	}

	// ── Per-repo summary table ────────────────────────────────────────────────
	sep := strings.Repeat("═", 97)
	rowSep := strings.Repeat("─", 97)
	totalCode, totalComments, totalBlank, totalSupported := 0, 0, 0, 0

	// Build into a strings.Builder so we can print AND optionally save to a file.
	var sb strings.Builder
	line := func(s string) {
		logger.Info(s)
		sb.WriteString(s + "\n")
	}

	line("")
	line(sep)
	line("  Per-Repository Results")
	line(fmt.Sprintf("  Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	line(sep)
	line(fmt.Sprintf("  %-45s  %14s  %14s  %10s  %10s",
		"Repository", "Supported LOC", "All-Lang LOC", "Comments", "Blank"))
	line(rowSep)
	for _, s := range summaries {
		line(fmt.Sprintf("  %-45s  %14d  %14d  %10d  %10d",
			s.FullName, s.SupportedCodeLines, s.CodeLineCount, s.CommentsLineCount, s.BlankLineCount))
		totalCode += s.CodeLineCount
		totalComments += s.CommentsLineCount
		totalBlank += s.BlankLineCount
		totalSupported += s.SupportedCodeLines
	}
	line(rowSep)
	line(fmt.Sprintf("  %-45s  %14d  %14d  %10d  %10d",
		"TOTAL", totalSupported, totalCode, totalComments, totalBlank))
	line(sep)
	line("")

	if args.SummaryFilePath != "" {
		if err := os.WriteFile(args.SummaryFilePath, []byte(sb.String()), 0644); err != nil {
			logger.Error("Could not write summary file: ", err)
		} else {
			logger.Info("Summary saved to ", args.SummaryFilePath)
		}
	}

	// ── CSV / HTML output ─────────────────────────────────────────────────────
	allFileResults = report.SortFileScanResults(allFileResults)
	aggregateTotal := report.CalculateTotalLineOfCode(allFileResults)

	writeCSVIfRequested(args.CsvFilePath, allFileResults, aggregateTotal)

	if args.HtmlReportsDirectoryPath != "" {
		for _, s := range summaries {
			repoHTMLDir := filepath.Join(args.HtmlReportsDirectoryPath, filepath.Base(s.FullName))
			writeHTMLIfRequested(repoHTMLDir, s.FullName, s.AllFileScanResults)
		}
	}

	printDisclaimer()
	printFooter(args.GithubOrg, totalSupported)

	fmt.Println(totalSupported)
}

// ──────────────────────────────────────────────────────────────────────────────
// Shared helpers
// ──────────────────────────────────────────────────────────────────────────────

func scanDirectory(dirPath string, ignorePatterns []string) []scanner.FileScanResults {
	filePaths := scanner.WalkDirectory(dirPath, ignorePatterns)
	var results []scanner.FileScanResults
	for _, fp := range filePaths {
		results = append(results, scanner.ScanFile(fp))
	}
	return results
}

func writeCSVIfRequested(csvFilePath string, fileScanResultsArr []scanner.FileScanResults, totalResult scanner.FileScanResults) {
	if csvFilePath == "" {
		return
	}
	records := report.ConvertFileResultsIntoRecords(fileScanResultsArr, totalResult)
	logger.Debug("Dumping results by file to ", csvFilePath)
	if err := report.WriteCsv(csvFilePath, records); err != nil {
		logger.Error("CSV reports will not be generated. Please check the file path and permissions.")
		logger.LogStackTraceAndExit(err)
	}
	logger.Info("Done! Results can be found ", csvFilePath)
}

func writeHTMLIfRequested(htmlDir, scanLabel string, fileScanResultsArr []scanner.FileScanResults) {
	if htmlDir == "" {
		return
	}
	if err := utilities.CreateDirectoryIfNotExists(htmlDir); err != nil {
		logger.LogStackTraceAndExit(err)
	}
	logger.Info("Dumping HTML report to ", htmlDir)
	fileNames, fileContents := report.GenerateHTMLReports(fileScanResultsArr)
	for i := range fileNames {
		report.WriteStringToFile(filepath.Join(htmlDir, fileNames[i]), fileContents[i])
	}
	report.DumpSVGs(htmlDir)
	logger.Info("Done! HTML report for ", scanLabel, " can be found in ", htmlDir)
}

func printDisclaimer() {
	logger.Info("")
	logger.Info("DISCLAIMER: This tool does not guarantee the accuracy of estimates for commercial conversations. Use industry standard tools for production accuracy requirements.")
	logger.Info("")
}

func printFooter(label string, supportedLOC int) {
	logger.Info("VERIFY THIS DOESN'T INCLUDE 3RD PARTY DEPENDENCIES, TEST CODE, AND OTHER NON-SOURCE CODE FILES FROM THIS ANALYSIS.")
	logger.Info("")
	logger.Info("https://docs.sonarsource.com/sonarqube-server/latest/server-upgrade-and-maintenance/monitoring/lines-of-code/ - LOC definitions.")
	logger.Info("")
	logger.Info("For detailed reporting, please use the --csv or --html options.")
	logger.Info("Total LOC for ", label, " for supported languages is ", supportedLOC)
}
