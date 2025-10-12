# GEMINI.md

## Project Overview

This project, LazyArchon, is a terminal-based task management TUI (Terminal User Interface) for Archon. It is built with Go and the [Bubble Tea](httpshttps://github.com/charmbracelet/bubbletea) library, providing a vim-like navigation experience for efficient task management directly from the terminal. The project is inspired by other popular terminal TUIs like [lazygit](https://github.com/jesseduffield/lazygit) and [lazydocker](https://github.com/jesseduffield/lazydocker).

The application fetches data from an Archon API server and displays it in a two-panel interface, with a task list on the left and a detailed view on the right. Users can browse projects, view task details, manage task statuses, and filter tasks by features.

## Building and Running

The project uses a `Makefile` to streamline the development process. Here are the key commands:

*   **Build the application:**
    ```bash
    make build
    ```
    This command builds the `lazyarchon` binary and places it in the `bin/` directory.

*   **Build and run the application:**
    ```bash
    make run
    ```

*   **Run the tests:**
    ```bash
    make test
    ```

*   **Run tests with coverage:**
    ```bash
    make test-coverage
    ```
    This will generate an HTML coverage report named `coverage.html`.

*   **Lint the code:**
    ```bash
    make lint
    ```
    This command will automatically fix formatting and simple linting issues.

*   **Run a comprehensive pre-commit check:**
    ```bash
    make check
    ```
    This will tidy dependencies, run the linter, and execute the test suite.

## Development Conventions

*   **Dependencies:** The project's dependencies are managed using Go modules. The `go.mod` file lists all the dependencies, including the Bubble Tea library and its related components for building the TUI.
*   **Linting:** The project uses `golangci-lint` for linting and `goimports` for formatting. The `Makefile` provides targets for running the linter and automatically fixing issues.
*   **Testing:** The project has a suite of tests that can be run with the `make test` command. The tests are located in the same packages as the code they are testing, following the `_test.go` naming convention.
*   **Architecture:** The project is structured with a clear separation of concerns. The UI is handled by the `internal/ui` package, which uses the Bubble Tea framework. The API client for interacting with the Archon server is located in the `internal/archon` package. Configuration is managed in the `internal/shared/config` package.
*   **Entry Point:** The main entry point of the application is `cmd/lazyarchon/main.go`. This file is responsible for parsing command-line flags, loading the configuration, initializing the UI model, and starting the Bubble Tea application.

# Archon Integration & Workflow

**CRITICAL: This project uses Archon for knowledge management, task tracking, and project organization.**

**MCP Project ID:** `6df238db-d06f-4ef8-807c-551bdaf3a19d`

## Core Archon Workflow Principles

### The Golden Rule: Task-Driven Development with Archon

**MANDATORY: Always complete the full Archon task cycle before any coding:**

1. **Check Current Task** → Review task details and requirements
2. **Research for Task** → Search relevant documentation and examples
3. **Implement the Task** → Write code based on research
4. **Update Task Status** → Move task from "todo" → "doing" → "review"
5. **Get Next Task** → Check for next priority task
6. **Repeat Cycle**

**Task Management Rules:**
- Update all actions to Archon
- Move tasks from "todo" → "doing" → "review" (not directly to complete)
- Maintain task descriptions and add implementation notes
- DO NOT MAKE ASSUMPTIONS - check project documentation for questions