# Project Requirements: CLI Command Finder

This document outlines the functional and non-functional requirements for the CLI command finder tool.

## 1. Functional Requirements

* **FR1: Core Command Search**: The tool must accept a natural language query from the user and search a local database to find and display relevant shell commands.
* **FR2: Customizable Call Command**: Users must be able to set a custom alias or shell function (e.g., `hey`, `miko`) to call the tool. The documentation must explain how to do this.
* **FR3: Interactive Command Builder**: For complex commands (e.g., `tar`, `ffmpeg`), the tool should offer an optional "wizard mode" that asks the user questions to build the command step-by-step.
* **FR4: Context-Aware Suggestions**: The tool should analyze the current working directory to provide more relevant command suggestions.
    * If in a Git repository, prioritize `git` commands.
    * If a `Dockerfile` is present, prioritize `docker` commands.
* **FR5: Personal Command Notebook**: The tool must have a `save` command allowing users to add their own custom commands to a personal database.
* **FR6: Pipeline/Recipe Support**: The command database should support storing and searching for common command pipelines (e.g., `grep ... | awk ...`).

---

## 2. Non-Functional Requirements

* **NFR1: Performance**: The tool must be fast, with search results appearing nearly instantaneously (<100ms).
* **NFR2: Resource Usage**: The application must be lightweight, with low CPU and memory consumption. It should not require a running background process.
* **NFR3: Compatibility**: The tool must be a single, statically linked binary that runs on major platforms (Linux, macOS, Windows via WSL).
* **NFR4: Usability**:
    * Installation should be simple (e.g., via a single binary download or a package manager like Homebrew).
    * The command database must use a human-readable format like YAML.
* **NFR5: Extensibility**: Users and contributors should find it easy to add new commands to the central database.