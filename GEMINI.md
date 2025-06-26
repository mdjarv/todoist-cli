# Gemini Project Overview: todoist-cli

This document provides a brief overview of the `todoist-cli` project to help Gemini assist with development.

## Project Purpose

This project is a Command Line Interface (CLI) tool for interacting with Todoist.

## API Documentation

The official Todoist API documentation can be found at [https://developer.todoist.com/api/v1](https://developer.todoist.com/api/v1).
The local API documentation is available at [docs/api.md](docs/api.md).

## Development Commands

### Running the application

To run the application for development purposes, use the following command:

```bash
go run main.go
```

### Building the application

To build the application binary, use the following command:

```bash
go build
```

### Running tests

To run the test suite, use the following command:

```bash
go test ./...
```

### Formatting code

This project uses `gofmt` for code formatting. To format the entire codebase, run:

```bash
gofmt -w .
```
