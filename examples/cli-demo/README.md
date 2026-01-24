# CLI and Shell Demo

This example demonstrates the use of the CLI and Shell standard library packages in Kukicha.

## Features

- **CLI Argument Parsing**: Parse command line arguments including commands, positional arguments, and flags
- **Shell Command Execution**: Execute shell commands safely with proper error handling
- **Integration**: Combine CLI parsing with shell command execution

## Usage

### Echo Command
```bash
# Simple echo
kukicha run examples/cli-demo/main.kuki echo "Hello World"

# With verbose flag
kukicha run examples/cli-demo/main.kuki echo --verbose "Hello World"
```

### List Directory
```bash
# List current directory
kukicha run examples/cli-demo/main.kuki ls

# List specific directory with verbose
kukicha run examples/cli-demo/main.kuki ls --verbose /tmp
```

### System Info
```bash
# Show system information
kukicha run examples/cli-demo/main.kuki info
```

## Code Structure

The example demonstrates:

1. **CLI Parsing**: Using `cli.Parse()` to extract commands, positional arguments, and flags
2. **Flag Handling**: Boolean flags with `cli.BoolFlag()`
3. **Shell Commands**: Running shell commands with `shell.RunSimple()`
4. **Error Handling**: Using `onerr` for robust error handling
5. **Command Routing**: Different behavior based on the command

## Key Functions Used

### CLI Package
- `cli.Parse()` - Parse command line arguments
- `cli.Command()` - Get the command name
- `cli.String()` - Get string arguments by index or name
- `cli.BoolFlag()` - Get boolean flag values
- `cli.Flag()` - Get flag values

### Shell Package
- `shell.RunSimple()` - Execute shell commands
- `shell.Which()` - Check if a command exists

## Implementation Notes

This example shows how to build a simple CLI tool that:
- Accepts different commands (echo, ls, info)
- Supports flags (--verbose)
- Executes shell commands safely
- Handles errors gracefully

The code is designed to be easy to understand and modify for your own CLI tools.