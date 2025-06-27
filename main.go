package main

import (
	"fmt"
	"go-cloc/logger"
	"go-cloc/report"
	"go-cloc/scanner"
	"go-cloc/utilities"
	"path/filepath"
)

func main() {
	// parse CLI arguments and store them in a struct
	args := utilities.ParseArgsFromCLI()

	// scan LOC for the directory
	logger.Info("Scanning ", args.LocalScanFilePath, "...")
	filePaths := scanner.WalkDirectory(args.LocalScanFilePath, args.IgnorePatterns)
	fileScanResultsArr := []scanner.FileScanResults{}
	for _, filePath := range filePaths {
		fileScanResultsArr = append(fileScanResultsArr, scanner.ScanFile(filePath))
	}

	logger.Debug("Calculating total LOC ...")

	// sort and calculate total LOC
	fileScanResultsArr = report.SortFileScanResults(fileScanResultsArr)
	repoTotalResult := report.CalculateTotalLineOfCode(fileScanResultsArr)
	repoTotalResultSupportedOnly := report.CalculateTotalLineOfCodeSupportedOnly(fileScanResultsArr)

	// Dump results by file in a csv
	if args.CsvFilePath != "" {
		// convert results into records for CSV or command line output
		records := report.ConvertFileResultsIntoRecords(fileScanResultsArr, repoTotalResult)
		logger.Debug("Dumping results by file to ", args.CsvFilePath)
		report.WriteCsv(args.CsvFilePath, records)
		logger.Info("Done! Results can be found ", args.CsvFilePath)
	}

	// Dump HTML reports
	if args.HtmlReportsDirectoryPath != "" {
		createHTMLDirectoryErr := utilities.CreateDirectoryIfNotExists(args.HtmlReportsDirectoryPath)
		if createHTMLDirectoryErr != nil {
			logger.Error("Error creating HTML reports directory: ", createHTMLDirectoryErr)
			logger.Error("HTML reports will not be generated. Please check the directory path and permissions.")
		} else {
			logger.Info("Dumping HTML report to ", args.HtmlReportsDirectoryPath)
			fileNames, fileContents := report.GenerateHTMLReports(fileScanResultsArr)

			for index, _ := range fileNames {
				fileName := fileNames[index]
				fileContent := fileContents[index]
				report.WriteStringToFile(filepath.Join(args.HtmlReportsDirectoryPath, fileName), fileContent)
			}
			report.DumpSVGs(args.HtmlReportsDirectoryPath)
			logger.Info("Done! HTML report for ", args.LocalScanFilePath, " can be found in ", args.HtmlReportsDirectoryPath)
		}
	}
	logger.Info("Printing total LOC for ALL languages ...")
	report.PrintResultsToCommandLine(repoTotalResult.CodeLineCount, repoTotalResult.CommentsLineCount, repoTotalResult.BlankLineCount)
	logger.Info("Printing LOC for SUPPORTED languages ...")
	report.PrintResultsToCommandLine(repoTotalResultSupportedOnly.CodeLineCount, repoTotalResultSupportedOnly.CommentsLineCount, repoTotalResultSupportedOnly.BlankLineCount)
	logger.Info("")
	logger.Info("VERIFY THIS DOESN'T INCLUDE 3RD PARTY DEPENDENCIES, TEST CODE, AND OTHER NON-SOURCE CODE FILES FROM THIS ANALYSIS.")
	logger.Info("")
	logger.Info("https://docs.sonarsource.com/sonarqube-server/latest/server-upgrade-and-maintenance/monitoring/lines-of-code/ - LOC definitions.")
	logger.Info("")
	logger.Info("For detailed reporting, please use the --csv or --html options. ")
	logger.Info("Total LOC for ", args.LocalScanFilePath, " for supported languages is ", repoTotalResultSupportedOnly.CodeLineCount)

	// Print the total LOC to standard output to make it easy for external tools to parse
	fmt.Println(repoTotalResultSupportedOnly.CodeLineCount)
}
