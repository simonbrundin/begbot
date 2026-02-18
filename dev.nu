#!/usr/bin/env nu

# -----------------------------------------------
# Begbot Development Server
# -----------------------------------------------

def get-local-ip [] {
    try {
        let ip_output = (ip route get 1.1.1.1 | complete)
        if $ip_output.exit_code == 0 {
            let ip = ($ip_output.stdout | str trim | parse -r 'src\s+(\S+)' | get capture0.0)
            $ip
        } else {
            let ips = (hostname -I | str trim | split row ' ')
            $ips.0
        }
    } catch {
        "127.0.0.1"
    }
}

# -----------------------------------------------
# Variabler
# -----------------------------------------------

let local_ip = (get-local-ip)
let backend_port = 8081
let frontend_port = 3000

# -----------------------------------------------
# Rensa gamla processer
# -----------------------------------------------

def kill-dev-ports [] {
    # Kill specific dev processes
    let patterns = ["air", "go run ./cmd/api/main.go", "npm run dev", "node.*nuxt", "node.*vite", "/tmp/begbot"]
    
    for $pattern in $patterns {
        try {
            let exit_code = (^bash -c $'pkill -9 -f "($pattern)" 2>/dev/null; echo $?' | str trim | into int)
            if $exit_code == 0 {
                print $"✓ Stängde processer för: ($pattern)"
            } else {
                print $"Inga processer hittades för: ($pattern)"
            }
        } catch {
            print $"Inga processer hittades för: ($pattern)"
        }
    }

    # Kill all Go and Node processes (for dev servers)
    try {
        ^bash -c 'pkill -9 node 2>/dev/null' | ignore
        print "✓ Stängde alla node-processer"
    } catch { }

    try {
        ^bash -c 'pkill -9 go 2>/dev/null' | ignore
        print "✓ Stängde alla go-processer"
    } catch { }

    # Kill by port for backend
    try {
        let pids = (^bash -c $'lsof -i :($backend_port) -t 2>/dev/null' | str trim)
        if $pids != "" {
            for $pid in ($pids | split row '\n') { kill -9 $pid | ignore }
            print $"✓ Stängde återstående processer på port ($backend_port)"
        }
    } catch { }

    # Kill by port for frontend
    try {
        let pids = (^bash -c $'lsof -i :($frontend_port) -t 2>/dev/null' | str trim)
        if $pids != "" {
            for $pid in ($pids | split row '\n') { kill -9 $pid | ignore }
            print $"✓ Stängde återstående processer på port ($frontend_port)"
        }
    } catch { }

    # Additional cleanup for any remaining on common dev ports
    for $port in [3000, 8080, 8081, 5173, 24678] {
        try {
            let pids = (^bash -c $'lsof -i :($port) -t 2>/dev/null' | str trim)
            if $pids != "" {
                for $pid in ($pids | split row '\n') { kill -9 $pid | ignore }
                print $"✓ Stängde processer på port ($port)"
            }
        } catch { }
    }
}

print "Städar upp gamla processer..."
kill-dev-ports

# -----------------------------------------------
# Välj läge
# -----------------------------------------------

print "\n🚀 Välj utvecklingsläge:\n"

let modes = [
    "all - Kör både backend och frontend",
    "backend - Kör endast Go backend",
    "frontend - Kör endast Nuxt frontend"
]

let selection = (try {
    $modes | str join "\n" | fzf --prompt="Mode: " --height=40% --reverse | str trim
} catch { "all" })

let mode = if ($selection | str contains "all") {
    "all"
} else if ($selection | str contains "frontend") {
    "frontend"
} else {
    "backend"
}

print $"\n✓ Startar i ($mode) mode...\n"

# -----------------------------------------------
# Starta servrar
# -----------------------------------------------

let backend_url = $"http://($local_ip):($backend_port)"
let frontend_url = $"http://($local_ip):($frontend_port)"
mut actual_frontend_url = $frontend_url

if $mode == "all" or $mode == "backend" {
    print $"🔧 Startar Go backend på port ($backend_port)..."
    cd /home/simon/repos/begbot
    ^bash -c "/home/simon/go/bin/air > /dev/null 2>&1 &"
    print $"✓ Backend startad: ($backend_url)"
}

if $mode == "all" or $mode == "frontend" {
    print $"🌐 Startar Nuxt frontend på port ($frontend_port)..."
    cd /home/simon/repos/begbot/frontend
    
    # Starta Nuxt och fånga output för att läsa faktisk port
    let log_file = "/tmp/nuxt-dev.log"
    ^bash -c $"API_BASE_URL='http://localhost:8081' npm run dev > ($log_file) 2>&1 &"
    
    # Vätta på att Nuxt startar och läs faktisk port
    mut actual_frontend_port = $frontend_port
    mut attempts = 0
    while $attempts < 30 {
        sleep 0.5sec
        let log_content = (try { open $log_file | str join "\n" } catch { "" })
        
        # Leta efter "Local:    http://localhost:XXXX" i output
        let port_match = ($log_content | parse -r 'Local:\s+http://[^:]+:(\d+)' | get capture0? | get 0?)
        if $port_match != null {
            $actual_frontend_port = ($port_match | into int)
            break
        }
        
        $attempts = $attempts + 1
    }
    
    $actual_frontend_url = $"http://($local_ip):($actual_frontend_port)"
    print $"✓ Frontend startad: ($actual_frontend_url)"
}

# -----------------------------------------------
# Notifiering
# -----------------------------------------------

# Bestäm vilken frontend-URL som ska visas
let final_frontend_url = if $mode == "all" or $mode == "frontend" {
    $actual_frontend_url
} else {
    $frontend_url
}

print "\n📡 Skickar notifiering..."
try {
    curl -d $"Backend: ($backend_url) | Frontend: ($final_frontend_url)" ntfy.sh/simonbrundin-dev-notification
} catch {
    print "Kunde inte skicka notifiering"
}

print "\n✅ Utvecklingsservrar körs!"
print $"   Backend: ($backend_url)"
print $"   Frontend: ($final_frontend_url)"
print "\nFör att stoppa: pkill -9 -f 'air' && pkill -9 -f 'npm run dev'"
