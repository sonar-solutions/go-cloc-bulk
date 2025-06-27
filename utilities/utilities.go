package utilities

import (
	"flag"
	"go-cloc/logger"
	"go-cloc/scanner"
	"os"
	"path/filepath"
	"strings"
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
	LogLevel                 string
	LocalScanFilePath        string
	IgnorePatterns           []string
	CsvFilePath              string
	HtmlReportsDirectoryPath string
}

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

	// Ensure the local scan file path is absolute
	args.LocalScanFilePath = CleanLocalFilePath(args.LocalScanFilePath)

	// Validate the local scan file path exists
	if _, err := os.Stat(args.LocalScanFilePath); os.IsNotExist(err) {
		logger.Error("The specified path does not exist: ", args.LocalScanFilePath)
		os.Exit(-1)
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
	// Define flags
	printLanguagesArg := flag.Bool("print-languages", false, "Prints out the supported languages, file suffixes, and comment configurations. Does not run the tool.")
	logLevelArg := flag.String("log-level", "INFO", "Log level - DEBUG, INFO, WARN, ERROR")
	ignoreFilePathArg := flag.String("ignore-file-path", "", "Path to your ignore file. Defines directories and files to exclude when scanning. Please see the README.md for how to format your ignore configuration")
	csvFilePathArg := flag.String("csv", "", "Path to dump results to a csv file, otherwise results are printed to standard out")
	htmlReportsDirectoryPathArg := flag.String("html", "", "Path to dump HTML reports into a specified directory, otherwise HTML reports are not generated. Note this directory must already exist.")
	overrideLanguageConfigFilePathArg := flag.String("override-languages", "", "Path to languages configuration to override the default configuration.")

	logger.Info("Parsing CLI arguments")
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

	// Ensure exactly one directory argument is provided
	if len(cliArgs) < 1 {
		logger.Error("Requires a path to the file or directory to scan as the first command line argument, ex: 'go-cloc file1.js'")
		os.Exit(-1)
	}

	// Set file path to scan
	localScanFilePath := CleanLocalFilePath(cliArgs[0])

	// parse ignore patterns
	ignorePatterns := []string{}
	if *ignoreFilePathArg != "" {
		logger.Debug("Parsing ignore-file ", *ignoreFilePathArg)
		ignorePatterns = scanner.ReadIgnoreFile(*ignoreFilePathArg)
		logger.Debug("Successfully read in the ignore-file ", *ignoreFilePathArg)
		logger.Debug("Ignore Patterns: ", ignorePatterns)
	}

	args := CLIArgs{
		LogLevel:                 *logLevelArg,
		LocalScanFilePath:        localScanFilePath,
		IgnorePatterns:           ignorePatterns,
		CsvFilePath:              *csvFilePathArg,
		HtmlReportsDirectoryPath: *htmlReportsDirectoryPathArg,
	}

	return args
}
