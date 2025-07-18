# Development Plan: Tasks

This document breaks down the project into small, actionable steps.

## Phase 1: Minimum Viable Product (MVP) 🚀

* [ ] **Task 1.1: Setup Project**: Initialize a Go module and project structure. Set up Git repository.
* [ ] **Task 1.2: Define Database Schema**: Create the final YAML structure for command entries.
* [ ] **Task 1.3: Research Command Sources**: Evaluate existing command databases (tldr, cheat.sh, navi) and determine the best sources for initial data population.
* [ ] **Task 1.4: Build Data Fetcher**: Create scripts to fetch command data from selected sources (e.g., tldr-pages GitHub repo, cheat.sh API).
* [ ] **Task 1.5: Create Data Converter**: Implement conversion logic to transform fetched data into our YAML schema format.
* [ ] **Task 1.6: Generate Initial Database**: Process and merge data from multiple sources, handle deduplication, and create the initial `commands.yml` file with 100+ commands.
* [ ] **Task 1.7: Implement CLI Parser**: Set up `cobra` to handle a basic search query.
* [ ] **Task 1.8: Build Basic Search**: Implement a simple, non-weighted search based on keyword matching.
* [ ] **Task 1.9: Display Results**: Format and print the search results to the console.
* [ ] **Task 1.10: Write Core README**: Document the purpose, installation (from source), and basic usage, including how to set up a shell alias.

---

## Phase 2: Enhancing Search & UX 🧠

* [ ] **Task 2.1: Text Processing**: Integrate stemming and stop-word removal into the search logic.
* [ ] **Task 2.2: Scoring Algorithm**: Implement a weighted scoring system to improve result relevance.
* [ ] **Task 2.3: Fuzzy Search**: Add a fuzzy search library to handle user typos gracefully.
* [ ] **Task 2.4: Personal Notebook (`save`)**:
    * Implement the `save` subcommand to append a new command to a personal YAML file.
    * Ensure the main search function loads commands from both the default and personal databases.

---

## Phase 3: Unique Features ✨

* [ ] **Task 3.1: Context Analyzer**:
    * Build the module to detect project types (`.git`, `Dockerfile`, etc.).
    * Integrate context into the scoring algorithm to boost relevant results.
* [ ] **Task 3.2: Interactive Builder**:
    * Design the state machine for the wizard.
    * Implement the builder for one or two complex commands like `tar` and `find`.

---

## Phase 4: Release & Polish 📦

* [ ] **Task 4.1: Testing**: Write unit tests for the search, scoring, and parsing modules.
* [ ] **Task 4.2: Build Automation**: Create a Makefile or build script to automate compilation for Linux, macOS, and Windows.
* [ ] **Task 4.3: Documentation**: Greatly expand the README with examples, GIFs, and details on all features.
* [ ] **Task 4.4: Release**: Tag v1.0.0 and create official releases on GitHub. Look into submitting to package managers like Homebrew.