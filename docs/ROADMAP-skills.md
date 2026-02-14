The Vision: A "Skill-First" File

In 2026, a developer shouldn't have to manually write a SKILL.md or a JSON schema. The Kukicha compiler should extract these from the code. see https://agentskills.io/specification

# weather_agent.kuki
petiole weather_tool

skill WeatherService
    description: "Provides real-time weather and climate data."
    version: "2.1.0"

    # This function is automatically exposed to the MCP server
    func GetForecast(city string, days int = 3) Forecast, error
        """Fetches a multi-day forecast for the specified city."""
        
        # Rough Edge: We need 'stdlib/http' to support 2026 A2A auth headers or should create a a2a stdlib that wraps https://github.com/a2aproject/a2a-go
        data := http.Get("https://api.weather.com/v3/{city}?d={days}")
            onerr return empty, error "Weather API unreachable"

        return data |> json.Unmarshal() as Forecast

1. Prototype: stdlib/mcp (The 2026 "Transport" Edge)

When an agent "calls" a tool, it uses stdio or SSE. If your Kukicha code uses print(), it will corrupt the JSON-RPC stream.

The Rough Edge: We need a Context-Aware Logger.

    Problem: print() goes to stdout.

    Solution: In "MCP Mode," print() should automatically redirect to stderr (which agents use for logging), while stdout is reserved for the protocol.

Code snippet

# stdlib/mcp logic (internal)
func Respond(id string, result any)
    # Uses Go 1.26 json/v2 for faster streaming
    response := map{
        "jsonrpc": "2.0",
        "id":      id,
        "result":  result,
    }
    # Write to RAW stdout, bypass the standard print()
    os.Stdout.Write(json.Marshal(response))

2. The skill Compiler Logic (Automatic SKILL.md)

The agentskills.io spec requires a specific folder structure. We should implement a kukicha pack command.

What kukicha pack does in 2026:

    Extracts Metadata: Takes the description and version from the skill block.

    Generates YAML: Creates the SKILL.md frontmatter.

    Reflects Schemas: Turns function parameters (like city string) into JSON Schema automatically.

    Compiles Binary: Build a tiny, static Go 1.26 binary into the scripts/ folder.

3. Identified "Rough Edges" for the Stdlib

JSON Typing	Go is too strict for LLMs.	Fluid Casting: args["days"] as int should handle the string "3" seamlessly.
Auth	Manual token management.	stdlib/a2a: Built-in support for "Agent-to-Agent" handshake tokens.

Error Feedback	Tracebacks are messy.	onerr explain: A way to send a "hint" back to the LLM (e.g., onerr explain "City names must be capitalized").

Concurrency	Goroutines are great but silent.	go with trace: Automatically attach a 2026 "OpenTelemetry" trace ID to every goroutine for agent debugging.

Lower-Level Stdlib Needs 

    stdlib/reflect: A way for Kukicha to inspect its own functions at runtime to serve the mcp/list_tools request.

    stdlib/json: see if we don't already we should check our implementation of the json/v2 package, which supports "omitempty" more cleanly—crucial for reducing token count in agent responses.

    stdlib/task: A higher-level abstraction for the A2A "Task" protocol, allowing an agent to "Pause" a script and wait for a human (the approve keyword). - maybe this is already in https://github.com/a2aproject/a2a-go

We want to allow the llm.kuki library to not just fail, but to tell the model exactly how to fix its prompt to succeed on the next try.
