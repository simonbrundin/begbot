# Standards: Port-based Process Kill

## Applicable Standards

### global/coding-style.md
- **Consistent Naming Conventions**: Use clear names like `kill-by-port`
- **Small, Focused Functions**: Keep the port-killing logic in one function
- **Remove Dead Code**: Replace old pkill calls with new function

### global/conventions.md
- **Consistent Project Structure**: Enhance existing dev.nu, don't create new files unnecessarily
- **Clear Documentation**: Comment the new function with its purpose

## Not Applicable
- configuration-structure.md (not configuration-related)
- currency-storage.md (not database-related)
- tech-stack.md (not technology selection)
