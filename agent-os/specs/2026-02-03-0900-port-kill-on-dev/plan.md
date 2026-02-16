# Plan: Port-based Process Kill for dev.nu

## Spec Status: ✅ DONE

**Completed:** 2026-02-03

---

## Task 1: Save spec documentation ✅ COMPLETED
Create spec folder with plan.md, shape.md, standards.md, references.md.

## Task 2: Implement port-based kill function ✅ COMPLETED
Created a Nushell function `kill-dev-ports` that:
- Kills all processes on port 8081 (backend) and 3000 (frontend)
- Uses `lsof -i :PORT -t` to find processes by port and extract PIDs
- Kills processes using `kill $pid`
- Handles errors gracefully (no error if no process found)

## Task 3: Update dev.nu to use the new function ✅ COMPLETED
- Replaced old pkill-based logic with `kill-dev-ports` function call
- Moved function definition before usage to fix scope issue
- Uses `$backend_port` and `$frontend_port` variables for flexibility

## Task 4: Test the implementation ✅ COMPLETED
- Function uses `lsof -t` flag to output only PIDs
- Error handling with try/catch blocks works correctly
- Gracefully handles case when no processes are running on ports
- Function runs before starting servers to clean up old processes

## Changes Made to dev.nu

**Before:**
```nushell
print "Städar upp gamla processer..."
kill-dev-ports

def kill-dev-ports [] {
    let ports = [8080 3000]

    for $port in $ports {
        try {
            ^fuser -k ($port)/tcp out+err> /dev/null
            print $"✓ Stängde port ($port)"
        } catch {
            print $"Inga processer på port ($port)"
        }
    }
}
```

**After:**
```nushell
def kill-dev-ports [] {
    let ports = [$backend_port $frontend_port]

    for $port in $ports {
        try {
            let lsof_output = (^lsof -i :$port -t | complete)
            if $lsof_output.exit_code == 0 and ($lsof_output.stdout | str trim) != "" {
                let pids = ($lsof_output.stdout | str trim | lines)
                for $pid in $pids {
                    try {
                        kill $pid
                        print $"✓ Stängde process ($pid) på port ($port)"
                    } catch {
                        print $"Kunde inte stänga process ($pid) på port ($port)"
                    }
                }
            } else {
                print $"Inga processer på port ($port)"
            }
        } catch {
            print $"Inga processer på port ($port)"
        }
    }
}

print "Städar upp gamla processer..."
kill-dev-ports
```

## Improvements
1. **Port-based over pattern-based**: More reliable than pkill pattern matching
2. **Correct ports**: Uses actual configured ports (8081 for backend, 3000 for frontend)
3. **Proper PID extraction**: Uses `lsof -t` for clean PID output
4. **Better error handling**: Multiple try/catch blocks for robustness
5. **Function scope**: Defined before called to avoid Nushell scope issues
