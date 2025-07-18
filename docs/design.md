# Design Document: CLI Command Finder

This document outlines the technical design and architecture for the tool.

## 1. High-Level Architecture

The tool will be a single, self-contained Command-Line Interface (CLI) application developed in **Go**. Go is chosen for its excellent performance, simple concurrency model, and ability to compile to a single, cross-platform binary.

The architecture will consist of the following core components:

* **CLI Parser**: Handles argument parsing, sub-commands (`search`, `save`, `wizard`), and flags. The `cobra` library will be used for this.
* **Command Database**:
    * Will use **YAML** for its human-readable syntax.
    * The database will be a collection of YAML files. A default one is shipped with the tool, and a second one can be created by the user (`~/.config/cmd-finder/personal.yml`).
    * **Schema**:
        ```yaml
        - command: "tar -czvf <archive_name>.tar.gz <directory>"
          description: "Create a compressed archive from a directory."
          keywords: [compress, archive, zip, tar, gzip]
          niche: "filesystem" # Optional: for context-aware search
          platform: [linux, macos] # Optional: for OS-specific commands
          pipeline: false # Is this a multi-command pipeline?
        ```
* **Search Engine**:
    1.  **Loader**: Reads and parses all YAML command files into memory on startup.
    2.  **Preprocessor**: Cleans user queries (lowercase, removes stop words, stemming).
    3.  **Scoring Algorithm**: Ranks results based on a weighted score from keyword matches, description matches, and context.
    4.  **Fuzzy Matching**: A library like `go-fuzzyfinder` can be integrated to handle typos.
* **Context Analyzer**: A module that checks for the existence of files/directories like `.git`, `Dockerfile`, `package.json` to influence the scoring algorithm.
* **Interactive Builder**: A state machine that walks a user through a series of prompts to construct a command string.

---

## 2. User Flow

1.  **Execution**: User types `hey <query>`.
2.  **Shell Expansion**: The shell converts `hey` to `cmd-finder`.
3.  **Parsing**: The `cobra` parser receives `<query>` as arguments.
4.  **Context Analysis**: The tool checks the current directory for context markers.
5.  **Search**: The Search Engine processes the query, scores all commands in the database (factoring in context), and identifies the top 3-5 results.
6.  **Display**: The results are printed to the console in a clean, readable format.
7.  **Interaction**: The user can then select a command to copy to the clipboard or execute.