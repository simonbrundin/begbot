# References: Port-based Process Kill

## Existing Implementation
- `dev.nu:34-37` - Current pkill-based cleanup logic

## Nushell Patterns
- Use `lsof -i :$port | complete` to find processes
- Parse PID from output using `str parse` or `split row`
- Use `kill $pid` to terminate

## Similar Code
- None existing in codebase for port-based killing

## External Resources
- Nushell `complete` command for running external commands
- `lsof` man page for port-based process discovery
