# CLI and Shell Demo

This example demonstrates the use of the CLI and Shell standard library packages in Kukicha using builder patterns.

## Features

- **CLI Argument Parsing**: Parse command line arguments using builder pattern with `cli.New()`, `cli.Arg()`, `cli.AddFlag()`
- **Shell Command Execution**: Execute shell commands safely using builder pattern with `shell.New()`, `shell.Execute()`
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

1. **CLI Builder Pattern**: Using `cli.New()` |> `cli.Arg()` |> `cli.AddFlag()` |> `cli.Action()` to define the CLI
2. **Running CLI App**: Using `cli.RunApp()` to execute the application
3. **Argument Access**: Using `cli.GetString()`, `cli.GetBool()` to retrieve arguments
4. **Shell Builder Pattern**: Using `shell.New()` to create commands
5. **Shell Execution**: Using `shell.Execute()` to run commands
6. **Result Handling**: Using `shell.Success()`, `shell.GetOutput()`, `shell.GetError()` to inspect results
7. **Error Handling**: Using `onerr` for robust error handling

## Key Functions Used

### CLI Package (Builder Pattern)
- `cli.New(name)` - Create a new CLI application
- `cli.Arg(name, description)` - Add an argument to the application
- `cli.AddFlag(name, description, default)` - Add a flag to the application
- `cli.Action(handler)` - Set the handler function
- `cli.RunApp(app)` - Run the CLI application
- `cli.GetString(args, name)` - Get string argument value
- `cli.GetBool(args, name)` - Get boolean flag value

### Shell Package (Builder Pattern)
- `shell.New(command, args...)` - Create a new shell command
- `shell.Dir(path)` - Set working directory (chainable)
- `shell.SetTimeout(seconds)` - Set timeout (chainable)
- `shell.Env(key, value)` - Add environment variable (chainable)
- `shell.Execute(cmd)` - Execute the command and return Result
- `shell.Success(result)` - Check if command succeeded
- `shell.GetOutput(result)` - Get stdout output
- `shell.GetError(result)` - Get stderr output
- `shell.Which(command)` - Check if a command exists

## Implementation Notes

This example shows how to build a simple CLI tool that:
- Accepts different commands (echo, ls, info)
- Supports flags (--verbose)
- Executes shell commands safely using the builder pattern
- Handles errors gracefully with proper result inspection

The code is designed to be easy to understand and modify for your own CLI tools.
