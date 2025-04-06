# Mermaid Diagram Generator and Validator

A Go tool for generating and validating Mermaid diagrams.

## Features

- Generate Mermaid diagrams from Go code files
- Generate diagrams for specific components (services, repositories, etc.)
- Create project-wide diagrams
- Validate Mermaid diagram syntax
- Fix syntax errors in Mermaid diagrams with multiple retry attempts
- Provide friendly explanations of syntax errors

## Installation

```bash
go build -o mm-gen ./cmd
```

## Usage

### Generating Diagrams

Generate a diagram from a Go file:
```bash
./mm-gen file [diagram-type] [file-path]
```

Generate a diagram for a component:
```bash
./mm-gen component [diagram-type] [component-type] [component-name]
```

Generate a project-wide diagram:
```bash
./mm-gen map [diagram-type]
```

### Validating Diagrams

Validate a Mermaid diagram:
```bash
./mm-gen validate [file-path]
```

Validate and attempt to fix syntax errors:
```bash
./mm-gen validate [file-path] --fix
```

Validate and explain syntax errors:
```bash
./mm-gen validate [file-path] --explain
```

### Advanced Options

Set the maximum number of retries for fixing diagrams:
```bash
./mm-gen validate [file-path] --fix --retries 5
```

Show verbose output during the fixing process:
```bash
./mm-gen validate [file-path] --fix --verbose
```

## Environment Variables

- `ANTHROPIC_API_KEY`: API key for Claude (required for fixing and explaining)
- `MERMAID_FIX_RETRIES`: Maximum number of retries for fixing diagrams (default: 3)

## Diagram Types

- `basic`: Basic diagram showing components and relationships
- `sequence`: Sequence diagram showing flow of execution
- `class`: Class diagram showing struct definitions and relationships
- `flowchart`: Flowchart diagram showing control flow
- `project`: Project-wide architecture diagram
- `config`: Configuration structure diagram
- `adapters`: Diagram showing inbound/outbound communications

## Examples

Validate a Mermaid diagram and fix it if there are errors:
```bash
./mm-gen validate examples/invalid.mmd --fix
```

Validate a diagram with 5 retry attempts and verbose output:
```bash
./mm-gen validate examples/complex.mmd --fix --retries 5 --verbose
```

Generate a class diagram for a service:
```bash
./mm-gen component class service DiagramService
``` 