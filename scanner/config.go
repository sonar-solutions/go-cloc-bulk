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
	"ActionScript": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".as", ".actionscript"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Abap": {
		LineComments:      []string{"\""},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".abap", ".ab4", ".flow"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Apex": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".cls", ".trigger"},
		FileNames:         []string{},
		IsSupported:       true,
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
	"COBOL": {
		LineComments:      []string{"*", "/"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".cbl", ".ccp", ".cob", ".cobol", ".cpy"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"C#": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".cs"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"CSS": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".css"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Golang": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".go"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"HTML": {
		LineComments:      []string{},
		MultiLineComments: [][]string{{"<!--", "-->"}},
		Extensions:        []string{".html", ".htm", ".cshtml", ".vbhtml", ".aspx", ".ascx", ".rhtml", ".erb", ".shtml", ".shtm", ".cmp"},
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
		Extensions:        []string{".js", ".jsx", ".jsp", ".jspx", ".jspf", ".mjs", ".vue"},
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
	"Kotlin": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".kt", ".kts"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Fortran": {
		LineComments:      []string{"!"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".f", ".for", ".f90", ".f95", ".f03", ".f08", ".f18", ".f20"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Flex": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".as"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"Groovy": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".groovy", ".gvy", ".gy"},
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
	"Perl": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"=begin", "=end"}},
		Extensions:        []string{".pl", ".pm", ".t", ".pod"},
		FileNames:         []string{},
		IsSupported:       false,
	},
	"PHP": {
		LineComments:      []string{"//", "#"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".php", ".php3", ".php4", ".php5", ".phtml", ".inc"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Objective-C": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".m"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Oracle PL/SQL": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".pkb"},
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
	"Python": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"\"\"\"", "\"\"\""}},
		Extensions:        []string{".py", ".python", ".ipynb"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"RPG": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".rpg"},
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
	"Scala": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".scala"},
		FileNames:         []string{},
		IsSupported:       true,
	},
	"Scss": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".scss"},
		FileNames:         []string{},
		IsSupported:       true,
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
	"Terraform": {
		LineComments:      []string{},
		MultiLineComments: [][]string{},
		Extensions:        []string{".tf"},
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
	"Docker": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".dockerfile"},
		FileNames:         []string{"Dockerfile"},
		IsSupported:       true,
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
		logger.LogStackTraceAndExit(err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		logger.LogStackTraceAndExit(err)
	}

	err = json.Unmarshal(byteValue, &Languages)
	if err != nil {
		logger.LogStackTraceAndExit(err)
	}
}
