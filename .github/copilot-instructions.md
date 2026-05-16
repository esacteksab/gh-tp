# GitHub Copilot Instructions for Go Projects

These instructions are based on the principles from ["Effective Go"](https://go.dev/doc/effective_go) and project-specific conventions. When generating, completing, or refactoring Go code, strictly adhere to the following guidelines.

## Formatting

- **Tooling:** Use `gofumpt` for code formatting instead of standard `go fmt`. Assume all code will be processed by `gofumpt` and ensure layout matches its stricter conventions.

## Commentary

- **Doc Comments:** Provide block comments or line comments (`//`) for every exported package, variable, constant, type, and function.
- **Sentences:** Make doc comments complete sentences that begin with the name of the element they describe.
- **Package Comments:** Every package should have a package comment (a block comment) placed immediately before the `package` clause.
- **Inline Comments for New Code:** Add concise inline comments for non-obvious logic in newly generated code. Explain intent and tradeoffs, not mechanical steps.
- **Mirror Existing Style:** Match the surrounding file’s comment tone, capitalization, sentence style, and verbosity when adding doc comments or inline comments.
- **New Unexported Functions:** Add short doc comments for newly generated unexported functions when behavior is not obvious from the name.
- **Test Comments:** For newly generated tests, include brief comments for scenario setup and purpose when the test flow is non-trivial.
- **Avoid Noise:** Do not add comments for self-explanatory lines or restate code literally.

## Naming Conventions

- **Package Names:** Keep package names short, concise, and lowercase. Avoid `under_scores` or `mixedCaps`. Avoid generic names like `util` or `common`.
- **Getters/Setters:** Do not put `Get` in getter names. If you have a field called `owner`, the getter method should be `Owner()` (capitalized for export), not `GetOwner()`.
- **Interface Names:** By convention, one-method interfaces are named by the method name plus an -er suffix (e.g., `Reader`, `Writer`, `Formatter`).
- **MixedCaps:** Use `MixedCaps` or `mixedCaps` rather than underscores to write multiword names.

## Control Structures

- **If Statements:** Avoid unnecessary parentheses around conditions.
- **Error Handling & Indentation:** Handle errors early and return immediately. Keep the "happy path" un-indented (guard clauses). Do not use `else` if the `if` block ends with a `return`.
- **For Loops:** Prefer `for ... range` when iterating over arrays, slices, maps, or channels.
- **Switch Statements:** Prefer `switch` statements over complex, chained `if-else` blocks.

## Functions & Methods

- **Multiple Returns:** Utilize Go's ability to return multiple values, especially for returning `(result, error)`.
- **Named Result Parameters:** Use named return parameters when it clarifies the code or when a `defer` statement needs to modify the returned values.
- **Defer:** Use `defer` statements for resource cleanup (e.g., closing files, unlocking mutexes) immediately after the resource is successfully acquired.
- **Pointers vs. Values for Receivers:** Use pointer receivers if the method modifies the receiver or if the receiver is a large struct. Be consistent within a given type.

## Data & Initialization

- **Allocation:** - Use `make` to initialize slices, maps, and channels.
- Use composite literals `T{}` over `new(T)` for initializing structs to clearly define fields.
- **Slices:** When initializing a slice with a known capacity, pre-allocate it using `make(type, length, capacity)` to avoid reallocation.
- **Maps:** Check for key existence using the comma-ok idiom: `val, ok := myMap[key]`.

## Concurrency

- **Philosophy:** "Don't communicate by sharing memory; share memory by communicating."
- **Channels & Goroutines:** Prefer channels for synchronizing state between goroutines over explicit locks (`sync.Mutex`), though `sync.Mutex` is perfectly acceptable for simple shared state.

## Error Handling

- **Return Errors:** Always return errors as the last argument in a function signature.
- **Custom Errors:** Use `fmt.Errorf` with the `%w` verb to wrap errors and preserve context.
- **Panic vs. Error:** Return standard `error` values for all expected failure conditions. Do not use `panic` for normal error handling; reserve it only for truly unrecoverable conditions (e.g., initialization failures).
