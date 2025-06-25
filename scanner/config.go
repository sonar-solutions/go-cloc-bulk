package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-cloc/logger"
	"io"
	"os"
)

type LanguageInfo struct {
	LineComments      []string   `json:"LineComments"`
	MultiLineComments [][]string `json:"MultiLineComments"`
	Extensions        []string   `json:"Extensions"`
	FileNames         []string   `json:"FileNames"`
	IsSupported       bool       `json:"IsSupported"`
}

var Languages = map[string]LanguageInfo{
	"Abap": {
		LineComments:      []string{"\""},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".abap", ".ab4", ".flow"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"ActionScript": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".as", ".actionscript"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Ada": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".ada", ".adb", ".ads", ".pad"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Apex": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".cls", ".trigger"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"AppleScript": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"(*", "*)"}},
		Extensions:        []string{".applescript"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"ASP": {
		LineComments:      []string{"//", "'"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".asa", ".ashx", ".asp", ".axd"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"ASP.NET": {
		LineComments:      []string{"//", "'"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".asax", ".ascx", ".asmx", ".aspx", ".master", ".sitemap", ".webinfo"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Assembly": {
		LineComments:      []string{";"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".asm", ".s", ".S", ".a51", ".nasm"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Bazel": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".bazel", ".bzl"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"C": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".c"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"C Header": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".h"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"C++": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".cpp", ".cc", ".cxx", ".c++"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"C++ Header": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".hh", ".hpp", ".hxx", ".h++", ".ipp"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"C#": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".cs", ".csx", ".cake", ".csharp", ".razor"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Clojure": {
		LineComments:      []string{";;"},
		MultiLineComments: [][]string{{"#_", "_#"}},
		Extensions:        []string{".clj", ".cljs", ".cljc", ".cljx"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"CoffeeScript": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"###", "###"}},
		Extensions:        []string{".coffee", ".cjsx", ".iced", ".cakefile"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"COBOL": {
		LineComments:      []string{"*", "/"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".cbl", ".ccp", ".cob", ".cobol", ".cpy"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"CSS": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".css", ".scss", ".sass", ".less"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Cucumber": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".feature"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Cuda": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".cu", ".cuh"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Dart": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".dart"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Delphi": {
		LineComments:      []string{"//", "//"},
		MultiLineComments: [][]string{{"{", "}"}, {"(*", "*)"}},
		Extensions:        []string{".dpr", ".dfm", ".dpk", ".dproj"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Docker": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".dockerfile"},
		FileNames:         []string{"Dockerfile"},
		IsSupported:       true,
	},
	"DOS Batch": {
		LineComments:      []string{"REM", "::"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".bat", ".cmd"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Elixir": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"=begin", "=end"}},
		Extensions:        []string{".ex", ".exs"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Erlang": {
		LineComments:      []string{"%"},
		MultiLineComments: [][]string{{"{-", "-}"}},
		Extensions:        []string{".app.src", ".emakefile", ".erl", ".hrl", ".rebar.config", ".rebar.config.lock", ".rebar.lock", ".xrl", ".yrl"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Flex": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".l", ".lex"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Fortran": {
		LineComments:      []string{"!"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".f", ".f77", ".f90", ".f95", ".for", ".ftn", ".pfo", ".f03", ".f08"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"F#": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"(*", "*)"}},
		Extensions:        []string{".fs", ".fsi", ".fsx"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Groovy": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".groovy", ".gant", ".grt", ".gtpl", ".gvy", ".jenkinsfile"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Golang": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".go"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Haskell": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"{-", "-}"}},
		Extensions:        []string{".hs", ".hsc", ".lhs"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"HTML": {
		LineComments:      []string{},
		MultiLineComments: [][]string{{"<!--", "-->"}},
		Extensions:        []string{".html", ".htm", ".cshtml", ".vbhtml", ".rhtml", ".erb", ".shtml", ".shtm", ".cmp"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Java": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".java", ".jav"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"JavaScript": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".js", ".mjs", ".cjs", ".es6", ".bones", ".jake", ".jakefile", ".jsb", ".jscad", ".jsfl", ".jsm", ".jss", ".njs", ".pac", ".sjs", ".ssjs", ".xsjs", ".xsjslib", ".vue", ".svelte"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"JCL": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".jcl", ".JCL"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"JSON": {
		// JSON does not have comments, but some parsers allow single-line comments
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".json", ".JSON", ".jsonl", ".jsonld", ".jsonc", ".json5"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"JSP": {
		LineComments:      []string{"//", "//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".jsp", ".jspx", ".tag", ".tagx", ".jspxf", ".jspf"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Kotlin": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".kt", ".kts", ".kotlin", ".ktm"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Lisp": {
		LineComments:      []string{";"},
		MultiLineComments: [][]string{{"#|", "|#"}},
		Extensions:        []string{".lisp", ".lsp", ".el", ".asd", ".cl", ".jl"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Lua": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"--[[", "]]"}},
		Extensions:        []string{".lua"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Make": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".mk", ".makefile", ".Makefile", ".gnumake", ".gnumakefile"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Markdown": {
		LineComments:      []string{"<!--", "-->"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".md", ".markdown", ".mdown", ".mdwn", ".mdx", ".mkd", ".mkdn", ".mkdown", ".ronn", ".workbook"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Objective-C": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".m", ".mm", ".objc", ".objcpp"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"OCaml": {
		LineComments:      []string{"//", "(*"},
		MultiLineComments: [][]string{{"(*", "*)"}},
		Extensions:        []string{".ml", ".mli", ".eliom", ".eliomi", ".mll", ".mly"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Pascal": {
		LineComments:      []string{"//", "//"},
		MultiLineComments: [][]string{{"{", "}"}, {"(*", "*)"}},
		Extensions:        []string{".pas", ".p", ".pp", ".ppr", ".lpr"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Perl": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"=begin", "=end"}},
		Extensions:        []string{".pl", ".pm", ".pl6", ".p6m", ".p6l", ".p6c", ".p6d", ".p6h", ".p6t", ".plx", ".ph", ".ph6"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"PHP": {
		LineComments:      []string{"//", "#"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".php", ".php3", ".php4", ".php5", "php7", ".phtml", ".inc", ".twig"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"PL/I": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".pl1"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"PL/SQL": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".pkb", ".pks", ".pls", ".plb", ".plsql", ".pql", ".pck", ".pckb", ".pcks"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"PowerShell": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"<#", "#>"}},
		Extensions:        []string{".ps1", ".psm1", ".psd1"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Python": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"\"\"\"", "\"\"\""}},
		Extensions:        []string{".py", ".python", ".ipynb"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"R": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".r", ".R", ".rmd"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"RPG": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".rpg", ".rpgle", ".rpglep", ".rpgp"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Ruby": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"=begin", "=end"}},
		Extensions:        []string{".rb"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Rust": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".rs"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Scala": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".scala", ".sc", ".sbt", ".kojo"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Shell Script": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".sh", ".bash", ".zsh", ".bashrc", ".bash_profile", ".zshrc", ".zprofile"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"SQL": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".sql"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Swift": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".swift"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Terraform": {
		LineComments:      []string{},
		MultiLineComments: [][]string{},
		Extensions:        []string{".tf", ".tfvars", ".hcl", ".nomad"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"TypeScript": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".ts", ".tsx"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"T-SQL": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".tsql"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Visual Basic .NET": {
		LineComments:      []string{"'"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".vb"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Visual Basic 6": {
		LineComments:      []string{"'"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".bas", ".frm"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"XML": {
		LineComments:      []string{"<!--"},
		MultiLineComments: [][]string{{"<!--", "-->"}},
		Extensions:        []string{".xml", ".XML", ".xsd", ".xsl"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"XHTML": {
		LineComments:      []string{"<!--"},
		MultiLineComments: [][]string{{"<!--", "-->"}},
		Extensions:        []string{".xhtml"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"YAML": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".yaml", ".yml"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Zig": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".zig"},
		FileNames:         []string{},
		IsSupported:       false,
	},
}

// Function to look up file information based on its extension
/*
@ext should match exactly as above, ".java" etc.
*/
func LookupByExtension(ext string) (string, LanguageInfo, bool) {
	for lang, info := range Languages {
		for _, languageExt := range info.Extensions {
			if languageExt == ext {
				return lang, info, true
			}
		}
	}
	return "", LanguageInfo{}, false
}

func LookupByFileName(fileName string) (string, LanguageInfo, bool) {
	for lang, info := range Languages {
		for _, languageFileName := range info.FileNames {
			if languageFileName == fileName {
				return lang, info, true
			}
		}
	}
	return "", LanguageInfo{}, false
}

func LookupLanguageInfo(languageName string) (LanguageInfo, bool) {
	info, found := Languages[languageName]
	if !found {
		return LanguageInfo{}, false
	}
	return info, true
}

func PrintLanguages() {
	logger.Info("Printing All Languages:")
	// Create a buffer to hold the JSON data
	var buf bytes.Buffer

	// Create a new JSON encoder and set SetEscapeHTML to false
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	// Encode the map to JSON
	if err := encoder.Encode(Languages); err != nil {
		logger.Error("Error encoding JSON: ", err)
		logger.LogStackTraceAndExit(err)
	}

	// Print the JSON string
	fmt.Println(buf.String())
}

// LoadLanguages reads the JSON file and overrides the default Languages map
func LoadLanguages(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		logger.Error("Error opening languages config file: ", err)
		logger.LogStackTraceAndExit(err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Error reading languages config file: ", err)
		logger.LogStackTraceAndExit(err)
	}

	// Create a temporary map and unmarshal into it
	var tempLanguages map[string]LanguageInfo
	err = json.Unmarshal(byteValue, &tempLanguages)
	if err != nil {
		logger.Error("Error unmarshalling languages config file: ", err)
		logger.LogStackTraceAndExit(err)
	}

	// Replace the global Languages map with the temporary one
	Languages = tempLanguages
}

func ValidateLanguagesConfig() error {
	fileSuffixToLanguage := make(map[string]string)

	for langName, langInfo := range Languages {
		for _, fileSuffix := range langInfo.Extensions {
			currLanguage, isFound := fileSuffixToLanguage[fileSuffix]
			if isFound { // check if the suffix already exists
				return fmt.Errorf("duplicate file suffix found: the file suffix '%s' for language '%s' is already defined in the configuration and is also defined for language '%s'", fileSuffix, langName, currLanguage)
			}
			fileSuffixToLanguage[fileSuffix] = langName // add the suffix to the map
		}
	}
	return nil
}
func GetLanguagesConfig() map[string]LanguageInfo {
	return Languages
}
