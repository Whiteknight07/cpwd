# cpwd

`cpwd` prints the current physical working directory and copies it to the macOS clipboard.

## Behavior

- On success, it prints `Copied to clipboard` and then the resolved path on the next line.
- On clipboard failure, it still prints the resolved path and exits with a non-zero status.
- It resolves symlinks so the copied value is the physical path, not the logical shell path.

## Install

### Homebrew

```bash
brew install whiteknight07/tap/cpwd
```

### From source

```bash
go install github.com/Whiteknight07/cpwd@latest
```

### Local development

```bash
go test ./...
go run .
```

## Example

```text
Copied to clipboard
/Users/stavan/projects/cpwd
```
