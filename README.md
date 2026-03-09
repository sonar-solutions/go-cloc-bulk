# go-cloc-bulk

> A fork of [sonar-solutions/go-cloc](https://github.com/sonar-solutions/go-cloc) extended with **automatic GitHub repository cloning** so you can count Lines of Code across an entire GitHub organisation — or any list of repositories — without cloning anything manually first.

[![Tutorial Video](https://img.youtube.com/vi/Yu2WXEgtMCc/0.jpg)](https://youtu.be/Yu2WXEgtMCc)

---

## Legal Disclaimer

This tool is designed to provide quick LOC estimates and does not guarantee accuracy for commercial or production use. For binding commercial conversations, use industry-standard tooling. This tool is provided for convenience purposes only.

---

## Installation

### Option 1: Install script (recommended — no Go required)

Detects your OS and CPU architecture, downloads the correct binary from GitHub Releases, and places it in your current directory.

```sh
curl -fsSL https://raw.githubusercontent.com/sonar-solutions/go-cloc-bulk/main/install.sh | bash
```

Move the binary somewhere on your `PATH` to use it from anywhere:

```sh
mv go-cloc /usr/local/bin/go-cloc
```

### Option 2: Manual download

1. Go to the [Releases page](https://github.com/sonar-solutions/go-cloc-bulk/releases/latest).
2. Download the ZIP for your platform (e.g. `go-cloc-darwin-arm64.zip`).
3. Unzip and run.

```sh
unzip go-cloc-darwin-arm64.zip
./go-cloc --help
```

### Option 3: Build from source

Requires Go 1.23+.

```sh
git clone https://github.com/sonar-solutions/go-cloc-bulk.git
cd go-cloc-bulk
go build -o go-cloc .
```

---

## Quick Start

### Scan a local directory

```sh
go-cloc ./my-project
go-cloc ./my-project --csv results.csv --html ./reports
```

### Scan a whole GitHub organisation (auto-clones every repo)

```sh
export GITHUB_TOKEN=ghp_your_token_here

go-cloc --github-org my-org --skip-archived --skip-forks --csv results.csv --summary summary.txt
```

### Scan an explicit list of repositories

```sh
# Inline list
go-cloc --github-repos "myorg/backend-api,myorg/frontend-app" --csv results.csv

# From a file (one owner/repo per line, # comments supported)
go-cloc --github-repos-file repos.txt --csv results.csv
```

---

## Configuration File

Instead of passing flags on every run, create a `go-cloc-config.json` file in the directory you run the tool from. It is **auto-detected** — no extra flag needed.

```sh
cp example_go-cloc-config.json go-cloc-config.json
# edit go-cloc-config.json with your settings, then just run:
go-cloc
```

> **Important:** Add `go-cloc-config.json` to your `.gitignore`. It may contain a GitHub token.

```sh
echo "go-cloc-config.json" >> .gitignore
```

You can also point to a config file explicitly:

```sh
go-cloc --config /path/to/my-config.json
```

### Example `go-cloc-config.json`

```json
{
  "github-org": "my-org",
  "github-token": "",
  "branch": "",
  "clone-dir": "/tmp/go-cloc-scan",
  "no-cleanup": true,
  "skip-archived": true,
  "skip-forks": false,
  "concurrency": 8,
  "csv": "./results.csv",
  "summary": "./summary.txt",
  "html": "",
  "log-level": "INFO",
  "ignore-file-path": "",
  "override-languages": ""
}
```

**Priority** (highest wins): CLI flag → config file → built-in default.

---

## GitHub Token

A token is required for private repositories and to avoid GitHub API rate limits on large organisations.

The token is resolved in this order:

1. `--github-token` CLI flag
2. `github-token` field in `go-cloc-config.json`
3. `$GITHUB_TOKEN` environment variable ← **recommended**

```sh
export GITHUB_TOKEN=ghp_your_token_here
```

Or add it to your shell profile (`~/.zshrc`, `~/.bashrc`) so it persists across sessions:

```sh
echo 'export GITHUB_TOKEN=ghp_your_token_here' >> ~/.zshrc
```

Generate a token at **GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)** with `repo` scope (read access is sufficient).

---

## Output

### Terminal + summary file

The per-repository table is printed to the terminal and, if `--summary` is set, saved to a text file with a timestamp.

```
═══════════════════════════════════════════════════════════════════════════════════════════════════
  Per-Repository Results
  Generated: 2026-03-09 17:25:48
═══════════════════════════════════════════════════════════════════════════════════════════════════
  Repository                                     Supported LOC    All-Lang LOC    Comments     Blank
───────────────────────────────────────────────────────────────────────────────────────────────────
  myorg/backend-api                                      34521           36200        4821      6102
  myorg/frontend-app                                     18230           18540        2103      3401
───────────────────────────────────────────────────────────────────────────────────────────────────
  TOTAL                                                  52751           54740        6924      9503
═══════════════════════════════════════════════════════════════════════════════════════════════════
```

### Understanding the columns

Every line in every file falls into exactly one of three buckets:

| Bucket | What it is |
|---|---|
| **Code** | Executable / declarative source lines |
| **Comments** | Lines that are purely a comment |
| **Blank** | Empty or whitespace-only lines |

These are mutually exclusive: `Code + Comments + Blank = Total lines in file`.

| Column | Scope | What's counted |
|---|---|---|
| **Supported LOC** | Sonar-supported languages only | Code lines only — **use this for Sonar licensing** |
| **All-Lang LOC** | Every file, every language | Code lines only |
| **Comments** | Every file | Comment lines only |
| **Blank** | Every file | Blank lines only |

The difference between **Supported LOC** and **All-Lang LOC** is your config/infra/data files (YAML, JSON, Markdown, shell scripts, etc.) — languages Sonar does not license against.

### CSV report

Provides a row per file, useful for filtering in Excel or similar tools:

```csv
filePath,languageName,isSupportedLanguage,blank,comment,code
/path/file1.js,JavaScript,true,10,100,1000
/path/file2.py,Python,true,10,100,1000
```

### HTML report

A navigable file-tree view of results. Generated per repository when `--html` is set.

---

## All Options

```sh
go-cloc --help
```

### Scan target (pick one)

| Flag | Description |
|---|---|
| `<path>` | Positional argument — local directory or file to scan |
| `--github-org` | GitHub organisation name. Clones and scans every repository. |
| `--github-repos` | Comma-separated list of `owner/repo` pairs to clone and scan. |
| `--github-repos-file` | Path to a plain-text file listing `owner/repo` pairs (one per line, `#` comments supported). |

### GitHub clone options

| Flag | Default | Description |
|---|---|---|
| `--github-token` | `$GITHUB_TOKEN` | Personal access token for private repos / higher rate limits. |
| `--branch` | Each repo's default | Branch to clone. Leave blank to use each repo's own default branch. |
| `--clone-dir` | OS temp dir | Directory where repos are cloned. |
| `--no-cleanup` | `false` | Keep cloned repos on disk after scanning. Re-runs skip the clone step. |
| `--skip-archived` | `true` | Skip archived repositories (org mode). |
| `--skip-forks` | `false` | Skip forked repositories (org mode). |
| `--concurrency` | `4` | Number of repositories to clone in parallel. |

### Output options

| Flag | Description |
|---|---|
| `--csv <path>` | Save per-file results to a CSV file. |
| `--html <dir>` | Save HTML reports to a directory. |
| `--summary <path>` | Save the per-repository summary table to a plain-text file. |

### General options

| Flag | Default | Description |
|---|---|---|
| `--config <path>` | `go-cloc-config.json` (auto) | Path to JSON config file. |
| `--ignore-file-path <path>` | | Path to an ignore file (see [Ignore Files](#ignore-files)). |
| `--override-languages <path>` | | Path to a custom language config JSON. |
| `--log-level` | `INFO` | `DEBUG`, `INFO`, `WARN`, or `ERROR`. |
| `--print-languages` | | Print all supported languages and exit. |
| `-v` | | Print version and exit. |

---

## Ignore Files

A plain-text file, one pattern per line, that excludes files and directories from scanning. Supports `*` wildcards.

```sh
# Ignore a whole directory
/path/to/node_modules/*

# Ignore by extension
*.min.js
*.lock

# Use it
go-cloc --github-org my-org --ignore-file-path ignore.txt
```

---

## Cloned Repository Behaviour

| Setting | Behaviour |
|---|---|
| `no-cleanup: false` (default) | Repos are deleted after each scan. Next run re-clones from scratch. |
| `no-cleanup: true` | Repos stay in `clone-dir`. Re-runs skip cloning — much faster. |

Set a fixed `clone-dir` together with `no-cleanup: true` for fast repeated runs:

```json
"clone-dir": "/tmp/go-cloc-scan",
"no-cleanup": true
```

---

## Repos File Format

```sh
# example_github_repos.txt
# Lines starting with '#' and blank lines are ignored.
myorg/backend-api
myorg/frontend-app
# myorg/this-one-is-commented-out
myorg/shared-library
```

---

## Scripting / CI Integration

The final line written to stdout is always the **Supported LOC** integer — easy to capture in scripts:

```sh
LOC=$(go-cloc --github-org my-org --skip-archived 2>/dev/null | tail -1)
echo "Total supported LOC: $LOC"
```

Non-zero exit code on failure.

---

## GitHub Actions

Two workflows are included in `.github/workflows/`:

### `cloc.yml` — scheduled scan

Runs on a cron schedule (or manually) and uploads CSV + HTML as workflow artifacts. No Docker required — runs on standard GitHub-hosted runners.

### `release.yml` — binary release pipeline

Triggered automatically when a version tag is pushed. Cross-compiles binaries for all platforms and attaches them to a GitHub Release.

```sh
git tag v2.1.0
git push origin v2.1.0
```

---

## Performance Benchmarks

```sh
# Scanning 1 Billion Lines of Code

# go-cloc finished in < 5s
time ./go-cloc one-billion-loc-test
3.9s user 0.72s system 93% cpu 4.976 total

# cloc finished in ~2.5 minutes
time cloc one-billion-loc-test
128.48s user 4.22s system 96% cpu 2:17.72 total
```

---

## Language Support

Run the following to see all supported languages, extensions, and whether each counts toward Sonar's supported LOC:

```sh
go-cloc --print-languages
```

Below is the default language configuration.

```json
{
  "Abap": {
    "LineComments": ["\""],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".abap", ".ab4", ".flow"],
    "FileNames": []
  },
  "ActionScript": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".as"],
    "FileNames": []
  },
  "Apex": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".cls", ".trigger"],
    "FileNames": []
  },
  "C": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".c"],
    "FileNames": []
  },
  "C Header": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".h"],
    "FileNames": []
  },
  "C#": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".cs"],
    "FileNames": []
  },
  "C++": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".cpp", ".cc", ".cxx", ".c++"],
    "FileNames": []
  },
  "C++ Header": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".hh", ".hpp", ".hxx", ".h++", ".ipp"],
    "FileNames": []
  },
  "COBOL": {
    "LineComments": ["*", "/"],
    "MultiLineComments": [],
    "Extensions": [".cbl", ".ccp", ".cob", ".cobol", ".cpy"],
    "FileNames": []
  },
  "CSS": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".css"],
    "FileNames": []
  },
  "Docker": {
    "LineComments": ["#"],
    "MultiLineComments": [],
    "Extensions": [".dockerfile"],
    "FileNames": ["Dockerfile"]
  },
  "Flex": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".as"],
    "FileNames": []
  },
  "Golang": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".go"],
    "FileNames": []
  },
  "HTML": {
    "LineComments": [],
    "MultiLineComments": [["<!--", "-->"]],
    "Extensions": [
      ".html",
      ".htm",
      ".cshtml",
      ".vbhtml",
      ".aspx",
      ".ascx",
      ".rhtml",
      ".erb",
      ".shtml",
      ".shtm",
      ".cmp"
    ],
    "FileNames": []
  },
  "JCL": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".jcl", ".JCL"],
    "FileNames": []
  },
  "Java": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".java", ".jav"],
    "FileNames": []
  },
  "JavaScript": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".js", ".jsx", ".jsp", ".jspx", ".jspf", ".mjs"],
    "FileNames": []
  },
  "Kotlin": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".kt", ".kts"],
    "FileNames": []
  },
  "Objective-C": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".m"],
    "FileNames": []
  },
  "Oracle PL/SQL": {
    "LineComments": ["--"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".pkb"],
    "FileNames": []
  },
  "PHP": {
    "LineComments": ["//", "#"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".php", ".php3", ".php4", ".php5", ".phtml", ".inc"],
    "FileNames": []
  },
  "PL/I": {
    "LineComments": ["--"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".pl1"],
    "FileNames": []
  },
  "Python": {
    "LineComments": ["#"],
    "MultiLineComments": [["\"\"\"", "\"\"\""]],
    "Extensions": [".py", ".python", ".ipynb"],
    "FileNames": []
  },
  "RPG": {
    "LineComments": ["#"],
    "MultiLineComments": [],
    "Extensions": [".rpg"],
    "FileNames": []
  },
  "Ruby": {
    "LineComments": ["#"],
    "MultiLineComments": [["=begin", "=end"]],
    "Extensions": [".rb"],
    "FileNames": []
  },
  "SQL": {
    "LineComments": ["--"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".sql"],
    "FileNames": []
  },
  "Scala": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".scala"],
    "FileNames": []
  },
  "Scss": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".scss"],
    "FileNames": []
  },
  "Swift": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".swift"],
    "FileNames": []
  },
  "T-SQL": {
    "LineComments": ["--"],
    "MultiLineComments": [],
    "Extensions": [".tsql"],
    "FileNames": []
  },
  "Terraform": {
    "LineComments": [],
    "MultiLineComments": [],
    "Extensions": [".tf"],
    "FileNames": []
  },
  "TypeScript": {
    "LineComments": ["//"],
    "MultiLineComments": [["/*", "*/"]],
    "Extensions": [".ts", ".tsx"],
    "FileNames": []
  },
  "Visual Basic .NET": {
    "LineComments": ["'"],
    "MultiLineComments": [],
    "Extensions": [".vb"],
    "FileNames": []
  },
  "Vue": {
    "LineComments": ["<!--"],
    "MultiLineComments": [["<!--", "-->"]],
    "Extensions": [".vue"],
    "FileNames": []
  },
  "XHTML": {
    "LineComments": ["<!--"],
    "MultiLineComments": [["<!--", "-->"]],
    "Extensions": [".xhtml"],
    "FileNames": []
  },
  "XML": {
    "LineComments": ["<!--"],
    "MultiLineComments": [["<!--", "-->"]],
    "Extensions": [".xml", ".XML", ".xsd", ".xsl"],
    "FileNames": []
  },
  "YAML": {
    "LineComments": ["#"],
    "MultiLineComments": [],
    "Extensions": [".yaml", ".yml"],
    "FileNames": []
  }
}

```
### Customising language support

Copy the JSON above, modify it to your needs, and pass it in via `--override-languages`:

```sh
go-cloc --github-org my-org --override-languages my-languages.json
```

Or set it in `go-cloc-config.json`:

```json
"override-languages": "./my-languages.json"
```