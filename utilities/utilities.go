package utilities

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-cloc/logger"
	"go-cloc/scanner"
	"os"
	"path/filepath"
	"strings"
)

const (
	VERSION string = "2.0.3" // Update this version when making releases)
)

// Modes
const (
	LOCAL       string = "Local"
	GITHUB      string = "GitHub"
	AZUREDEVOPS string = "AzureDevOps"
	GITLAB      string = "GitLab"
	BITBUCKET   string = "Bitbucket"
)

type CLIArgs struct {
	// ── Core scan options ────────────────────────────────────────────────────
	LogLevel                 string
	LocalScanFilePath        string // path to scan in local mode
	IgnorePatterns           []string
	CsvFilePath              string
	HtmlReportsDirectoryPath string
	SummaryFilePath          string // --summary

	// ── GitHub auto-clone options ────────────────────────────────────────────
	// Set GithubOrg to scan every repository in a GitHub organisation.
	// Set GithubRepos to scan an explicit list of "owner/repo" pairs.
	// Set GithubReposFile to read the list from a file (one "owner/repo" per line).
	// All three flags can be combined; duplicates are deduplicated automatically.
	GithubOrg       string   // --github-org
	GithubRepos     []string // --github-repos (comma-separated)
	GithubReposFile string   // --github-repos-file
	GithubToken     string   // --github-token  (falls back to $GITHUB_TOKEN)
	Branch          string   // --branch  (default: each repo's default branch)
	CloneDir        string   // --clone-dir     (default: OS temp dir)
	NoCleanup       bool     // --no-cleanup    keep cloned repos after scan
	SkipArchived    bool     // --skip-archived (default true)
	SkipForks       bool     // --skip-forks
	Concurrency     int      // --concurrency   parallel clone workers (default 4)
}

// defaultScanConfigFileName is automatically loaded when present in the working
// directory, without needing to pass --config on every invocation.
const defaultScanConfigFileName = "go-cloc-config.json"

// ScanConfig mirrors every CLI flag so the same options can be persisted in a
// JSON file.  CLI flags always take precedence over config file values, which
// in turn take precedence over built-in defaults.
//
// Generate a starter file with all options:
//   cp example_go-cloc-config.json go-cloc-config.json
type ScanConfig struct {
	LogLevel          string `json:"log-level"`
	IgnoreFilePath    string `json:"ignore-file-path"`
	CsvFilePath       string `json:"csv"`
	HtmlDir           string `json:"html"`
	SummaryFilePath   string `json:"summary"`
	OverrideLanguages string `json:"override-languages"`
	GithubOrg         string `json:"github-org"`
	GithubRepos       string `json:"github-repos"`
	GithubReposFile   string `json:"github-repos-file"`
	GithubToken       string `json:"github-token"`
	Branch            string `json:"branch"`
	CloneDir          string `json:"clone-dir"`
	NoCleanup         bool   `json:"no-cleanup"`
	// Pointer so we can distinguish an explicit false from "not set"
	SkipArchived *bool `json:"skip-archived"`
	SkipForks    bool  `json:"skip-forks"`
	Concurrency  int   `json:"concurrency"`
}

// findConfigPath does a quick pre-scan of os.Args looking for --config / -config
// before the flag package is initialized, so we can load the file and use its
// values as flag defaults.
func findConfigPath() string {
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--config" || arg == "-config" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
		if strings.HasPrefix(arg, "-config=") {
			return strings.TrimPrefix(arg, "-config=")
		}
	}
	return ""
}

func loadScanConfig(path string) (ScanConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ScanConfig{}, err
	}
	var cfg ScanConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ScanConfig{}, fmt.Errorf("parsing config file %q: %w", path, err)
	}
	return cfg, nil
}

func boolPtr(b bool) *bool { return &b }

func CleanLocalFilePath(targetPath string) string {
	logger.Debug("CleanLocalFilePath targetPath before: '", targetPath, "'")
	targetPath = filepath.Clean(targetPath)
	// On windows this may be needed if spaces are in the file path
	targetPath = strings.TrimSuffix(targetPath, "\"")
	logger.Debug("CleanLocalFilePath targetPath after: '", targetPath, "'")
	return targetPath
}

func GetArgsFromCLI() CLIArgs {
	// Parse command line arguments
	args := ParseArgsFromCLI()

	// Local-mode validation: ensure the path exists.
	// Skip when GitHub mode is active (no local path expected).
	if args.GithubOrg == "" && len(args.GithubRepos) == 0 {
		args.LocalScanFilePath = CleanLocalFilePath(args.LocalScanFilePath)
		if _, err := os.Stat(args.LocalScanFilePath); os.IsNotExist(err) {
			logger.Error("The specified path does not exist: ", args.LocalScanFilePath)
			os.Exit(-1)
		}
	}

	return args
}

// create directory if it doesn't exist
func CreateDirectoryIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		logger.Debug("Creating directory: ", dirPath)
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			logger.Error("Error creating directory: ", err)
			return err
		}
	} else {
		logger.Debug("Directory already exists: ", dirPath)
	}
	return nil
}

func ParseArgsFromCLI() CLIArgs {
	// ── Load config file before defining flags so its values become defaults ──
	// Priority (highest to lowest): CLI flag → config file → built-in default
	cfg := ScanConfig{
		LogLevel:     "INFO",
		Concurrency:  4,
		SkipArchived: boolPtr(true),
	}

	configPath := findConfigPath()
	if configPath == "" {
		if _, err := os.Stat(defaultScanConfigFileName); err == nil {
			configPath = defaultScanConfigFileName
		}
	}
	if configPath != "" {
		loaded, err := loadScanConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not load config file %q: %v\n", configPath, err)
		} else {
			fmt.Fprintf(os.Stdout, "[INFO] Loaded config from %s\n", configPath)
			if loaded.LogLevel != "" {
				cfg.LogLevel = loaded.LogLevel
			}
			if loaded.IgnoreFilePath != "" {
				cfg.IgnoreFilePath = loaded.IgnoreFilePath
			}
			if loaded.CsvFilePath != "" {
				cfg.CsvFilePath = loaded.CsvFilePath
			}
			if loaded.HtmlDir != "" {
				cfg.HtmlDir = loaded.HtmlDir
			}
			if loaded.SummaryFilePath != "" {
				cfg.SummaryFilePath = loaded.SummaryFilePath
			}
			if loaded.OverrideLanguages != "" {
				cfg.OverrideLanguages = loaded.OverrideLanguages
			}
			if loaded.GithubOrg != "" {
				cfg.GithubOrg = loaded.GithubOrg
			}
			if loaded.GithubRepos != "" {
				cfg.GithubRepos = loaded.GithubRepos
			}
			if loaded.GithubReposFile != "" {
				cfg.GithubReposFile = loaded.GithubReposFile
			}
			if loaded.GithubToken != "" {
				cfg.GithubToken = loaded.GithubToken
			}
			if loaded.Branch != "" {
				cfg.Branch = loaded.Branch
			}
			if loaded.CloneDir != "" {
				cfg.CloneDir = loaded.CloneDir
			}
			if loaded.NoCleanup {
				cfg.NoCleanup = true
			}
			if loaded.SkipArchived != nil {
				cfg.SkipArchived = loaded.SkipArchived
			}
			if loaded.SkipForks {
				cfg.SkipForks = true
			}
			if loaded.Concurrency > 0 {
				cfg.Concurrency = loaded.Concurrency
			}
		}
	}

	// ── Flag definitions — config values are the defaults, CLI overrides them ─
	versionArg := flag.Bool("v", false, "Show version information")
	printLanguagesArg := flag.Bool("print-languages", false, "Prints out the supported languages, file suffixes, and comment configurations. Does not run the tool.")
	_ = flag.String("config", configPath, "Path to a JSON config file. If not set, go-cloc-config.json in the current directory is used automatically if present.")
	logLevelArg := flag.String("log-level", cfg.LogLevel, "Log level - DEBUG, INFO, WARN, ERROR")
	ignoreFilePathArg := flag.String("ignore-file-path", cfg.IgnoreFilePath, "Path to your ignore file. Defines directories and files to exclude when scanning. Please see the README.md for how to format your ignore configuration")
	csvFilePathArg := flag.String("csv", cfg.CsvFilePath, "Path to dump results to a csv file, otherwise results are printed to standard out")
	htmlReportsDirectoryPathArg := flag.String("html", cfg.HtmlDir, "Path to dump HTML reports into a specified directory, otherwise HTML reports are not generated. Note this directory must already exist.")
	summaryFilePathArg := flag.String("summary", cfg.SummaryFilePath, "Path to write the per-repository summary table as a plain-text file, e.g. ./summary.txt")
	overrideLanguageConfigFilePathArg := flag.String("override-languages", cfg.OverrideLanguages, "Path to languages configuration to override the default configuration.")

	// ── GitHub auto-clone flags ───────────────────────────────────────────────
	githubOrgArg := flag.String("github-org", cfg.GithubOrg, "GitHub organisation name. Clones and scans every repository in the org.")
	githubReposArg := flag.String("github-repos", cfg.GithubRepos, "Comma-separated list of repositories to clone and scan, e.g. 'owner/repo1,owner/repo2'.")
	githubReposFileArg := flag.String("github-repos-file", cfg.GithubReposFile, "Path to a plain-text file listing repositories to scan, one 'owner/repo' per line. Lines starting with '#' are treated as comments.")
	githubTokenArg := flag.String("github-token", cfg.GithubToken, "GitHub personal access token (PAT). Falls back to $GITHUB_TOKEN environment variable.")
	branchArg := flag.String("branch", cfg.Branch, "Branch to clone for every repository. Defaults to each repository's own default branch (main, master, etc.) when not set.")
	cloneDirArg := flag.String("clone-dir", cfg.CloneDir, "Directory in which repositories are cloned. Defaults to a temporary directory that is removed after the scan.")
	noCleanupArg := flag.Bool("no-cleanup", cfg.NoCleanup, "Do not delete cloned repositories after scanning.")
	skipArchivedArg := flag.Bool("skip-archived", *cfg.SkipArchived, "Skip archived repositories (applies to --github-org).")
	skipForksArg := flag.Bool("skip-forks", cfg.SkipForks, "Skip forked repositories (applies to --github-org).")
	concurrencyArg := flag.Int("concurrency", cfg.Concurrency, "Number of repositories to clone in parallel.")

	logger.Debug("Parsing CLI arguments")
	// Parse flags
	flag.Parse()
	// Parse non-flag arguments
	cliArgs := flag.Args()
	logger.Debug("Non-flag arguments: ", cliArgs)
	// If there are any non-flag arguments, we assume the first one is the directory to scan
	if len(cliArgs) > 0 {
		// Parse any remaining flags after the first non-flag argument
		flag.CommandLine.Parse(cliArgs[1:])
	}

	if *versionArg {
		fmt.Println("v" + VERSION)
		os.Exit(0)
	}

	// Set log level immediately after parsing
	logger.Info("Setting Log Level to " + *logLevelArg)
	logger.SetLogLevel(logger.ConvertStringToLogLevel(*logLevelArg))
	logger.SetOutput(os.Stdout)

	// Print out the command line arguments for debugging purposes
	logger.Debug("Command Line Arguments:")
	logger.Debug("print-languages: ", *printLanguagesArg)
	logger.Debug("log-level: ", *logLevelArg)
	logger.Debug("ignore-file-path: ", *ignoreFilePathArg)
	logger.Debug("csv-file-path: ", *csvFilePathArg)
	logger.Debug("html-reports-directory-path: ", *htmlReportsDirectoryPathArg)
	logger.Debug("override-language-config-file-path: ", *overrideLanguageConfigFilePathArg)
	logger.Debug("github-org: ", *githubOrgArg)
	logger.Debug("github-repos: ", *githubReposArg)
	logger.Debug("clone-dir: ", *cloneDirArg)
	logger.Debug("no-cleanup: ", *noCleanupArg)
	logger.Debug("skip-archived: ", *skipArchivedArg)
	logger.Debug("skip-forks: ", *skipForksArg)
	logger.Debug("concurrency: ", *concurrencyArg)
	logger.Debug("github-repos-file: ", *githubReposFileArg)

	// Override languages config if specified
	if *overrideLanguageConfigFilePathArg != "" {
		logger.Info("Overriding default languages with ", *overrideLanguageConfigFilePathArg)
		scanner.LoadLanguages(*overrideLanguageConfigFilePathArg)
		scanner.ValidateLanguagesConfig()
		logger.Debug("Successfully loaded languages configuration from ", *overrideLanguageConfigFilePathArg)
	}

	// Handle short-circuit flag: print languages and exit
	if *printLanguagesArg {
		scanner.PrintLanguages()
		os.Exit(0)
	}

	// ── Resolve GitHub token ──────────────────────────────────────────────────
	githubToken := *githubTokenArg
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
	}

	// ── Parse explicit repo list ──────────────────────────────────────────────
	var githubRepos []string
	if *githubReposArg != "" {
		for _, r := range strings.Split(*githubReposArg, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				githubRepos = append(githubRepos, r)
			}
		}
	}
	if *githubReposFileArg != "" {
		fileRepos, err := readReposFile(*githubReposFileArg)
		if err != nil {
			logger.Error("Failed to read --github-repos-file: ", err)
			os.Exit(-1)
		}
		githubRepos = append(githubRepos, fileRepos...)
	}

	// ── Determine operating mode ──────────────────────────────────────────────
	isGithubMode := *githubOrgArg != "" || len(githubRepos) > 0

	// In local mode a positional path argument is required
	localScanFilePath := ""
	if !isGithubMode {
		if len(cliArgs) < 1 {
			logger.Error("Requires a path to the file or directory to scan as the first command line argument, ex: 'go-cloc .'")
			logger.Error("Or use --github-org / --github-repos to clone and scan GitHub repositories automatically.")
			os.Exit(-1)
		}
		localScanFilePath = CleanLocalFilePath(cliArgs[0])
	} else if len(cliArgs) > 0 {
		// A positional arg alongside GitHub flags is treated as an additional
		// local path — interpret it as the clone directory if --clone-dir was
		// not set.
		if *cloneDirArg == "" {
			*cloneDirArg = CleanLocalFilePath(cliArgs[0])
		}
	}

	// parse ignore patterns
	ignorePatterns := []string{}
	if *ignoreFilePathArg != "" {
		logger.Debug("Parsing ignore-file ", *ignoreFilePathArg)
		ignorePatterns = scanner.ReadIgnoreFile(*ignoreFilePathArg)
		logger.Debug("Successfully read in the ignore-file ", *ignoreFilePathArg)
		logger.Debug("Ignore Patterns: ", ignorePatterns)
	}

	return CLIArgs{
		LogLevel:                 *logLevelArg,
		LocalScanFilePath:        localScanFilePath,
		IgnorePatterns:           ignorePatterns,
		CsvFilePath:              *csvFilePathArg,
		HtmlReportsDirectoryPath: *htmlReportsDirectoryPathArg,
		SummaryFilePath:          *summaryFilePathArg,
		GithubOrg:                *githubOrgArg,
		GithubRepos:              githubRepos,
		GithubReposFile:          *githubReposFileArg,
		GithubToken:              githubToken,
		Branch:                   *branchArg,
		CloneDir:                 *cloneDirArg,
		NoCleanup:                *noCleanupArg,
		SkipArchived:             *skipArchivedArg,
		SkipForks:                *skipForksArg,
		Concurrency:              *concurrencyArg,
	}
}

// readReposFile reads a plain-text file and returns the non-empty, non-comment
// lines as a slice of "owner/repo" strings.
func readReposFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var repos []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		repos = append(repos, line)
	}
	return repos, nil
}
