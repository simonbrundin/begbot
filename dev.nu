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

# Hitta lediga portar för backend och frontend så varje worktree kan köra egna dev-servrar
def find-free-port [] {
    try {
        let out = (^bash -c "python3 -c 'import socket; s=socket.socket(); s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1); s.bind((\"127.0.0.1\",0)); print(s.getsockname()[1]); s.close()'" | str trim)
        ($out | into int)
    } catch { 0 }
}

# Gör variabler mutabla så vi kan reasignera utan skuggning
# Lagra per-worktree portar i .dev.env för stabilitet över restarts
let devenv_file = ".dev.env"

# Hantera --reset flagg (ta bort .dev.env om användaren begär)
mut reset_requested = false
try {
    for $arg in $nu.args {
        if $arg == "--reset" { $reset_requested = true }
    }
} catch { }

if $reset_requested {
    try { ^bash -c $"rm -f ($devenv_file)" } catch { }
    print "⚠️ .dev.env raderad via --reset"
}

# Hantera --no-install flagg (förhindra automatisk installation av socat)
mut no_install_requested = false
try {
    for $arg in $nu.args {
        if $arg == "--no-install" { $no_install_requested = true }
    }
} catch { }

# Hantera --help / -h
mut help_requested = false
try {
    for $arg in $nu.args {
        if ($arg == "--help") { $help_requested = true }
        if ($arg == "-h") { $help_requested = true }
    }
} catch { }

if $help_requested {
    print "dev.nu - utvecklingsserver för worktrees"
    print ""
    print "Usage: ./dev.nu [options]"
    print ""
    print "Options:"
    print "  --help, -h        Visa denna hjälpskärm"
    print "  --reset           Radera .dev.env och välj nya portar"
    print "  --no-install      Hoppa över automatisk installation av socat"
    print ""
    print "Behavior: Skriptet väljer lediga portar per worktree, sparar dem i .dev.env (ignorerad av git),"
    print "and startar backend (air) och frontend (npm run dev). If socat is available it uses a proxy to"
    print "bind public ports and forward to internal ports to avoid race conditions. Production config"
    print "is not affected."
    exit 0
}

# Läs in befintlig .dev.env om den finns
mut backend_port = 0
mut frontend_port = 0
mut backend_internal_port = 0
mut frontend_internal_port = 0

# Read .dev.env using bash cat for robustness
let raw = (try { ^bash -c $"cat ($devenv_file) 2>/dev/null || true" | str trim } catch { "" })
if $raw != "" {
    for $line in ($raw | lines) {
        if ($line | str starts-with "BACKEND_PORT=")          { $backend_port          = (($line | split row '=' | get 1) | into int) }
        if ($line | str starts-with "FRONTEND_PORT=")         { $frontend_port         = (($line | split row '=' | get 1) | into int) }
        if ($line | str starts-with "BACKEND_INTERNAL_PORT=") { $backend_internal_port  = (($line | split row '=' | get 1) | into int) }
        if ($line | str starts-with "FRONTEND_INTERNAL_PORT=") { $frontend_internal_port = (($line | split row '=' | get 1) | into int) }
    }
}

# Om inga portar hittades i fil, hitta nya
if $backend_port == 0          { $backend_port          = (find-free-port) }
if $frontend_port == 0         { $frontend_port         = (find-free-port) }
if $backend_internal_port == 0  { $backend_internal_port  = (find-free-port) }
if $frontend_internal_port == 0 { $frontend_internal_port = (find-free-port) }

# Kontrollera om socat finns för bind-and-hold proxy
let have_socat = (try { (^bash -c "command -v socat >/dev/null && echo yes || echo no" | str trim) } catch { "no" })

# Om socat saknas: visa instruktion (installation kan köras manuellt). Vi undviker komplex
# interaktiv installationslogik här för stabilitet i olika terminalmiljöer.
if $have_socat == "no" {
    try { ^bash -c "notify-send -u normal 'OpenCode: Question' 'dev.nu: socat saknas. Installera socat för bind-and-hold eller kör med --no-install'" } catch { }

    if not $no_install_requested {
        print "⚠️ socat saknas. För att få mest robust beteende, installera socat."
        print "Manuella kommandon (välj din distro):"
        print "  Debian/Ubuntu: sudo apt-get update && sudo apt-get install -y socat"
        print "  Fedora: sudo dnf install -y socat"
        print "  Arch: sudo pacman -Syu socat"
        print "  Alpine: sudo apk add socat"
        print "  macOS (Homebrew): brew install socat"
        print "Eller kör: ./dev.nu --no-install för att hoppa över automatisk installation."
    } else {
        print "--no-install satt: hoppar över installation av socat. Fortsätter utan bind-and-hold."
    }
}

# -----------------------------------------------
# Rensa gamla processer (port-baserat, ej pattern)
# -----------------------------------------------

def kill-dev-ports [ports: list<int>] {
    for $p in $ports {
        if $p == 0 { continue }
        try {
            let exit_code = (^bash -c $"lsof -i :($p) -t 2>/dev/null | xargs -r kill -9; echo $?" | str trim | into int)
            if $exit_code == 0 {
                print $"✓ Stängde processer på port ($p)"
            }
        } catch { }
    }
}

print "Städar upp gamla processer..."
kill-dev-ports [$backend_port, $frontend_port, $backend_internal_port, $frontend_internal_port]

# Generera nya interna portar efter cleanup
$backend_internal_port  = (find-free-port)
$frontend_internal_port = (find-free-port)

# Skriv alla portar till .dev.env
try {
    let file_content = $"BACKEND_PORT=($backend_port)\nFRONTEND_PORT=($frontend_port)\nBACKEND_INTERNAL_PORT=($backend_internal_port)\nFRONTEND_INTERNAL_PORT=($frontend_internal_port)\n"
    $file_content | save -f $devenv_file
} catch { print "Kunde inte skriva .dev.env" }

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
} else if ($selection | str contains "backend") {
    "backend"
} else {
    "all"
}

print $"\n✓ Startar i ($mode) mode...\n"

# -----------------------------------------------
# Starta servrar
# -----------------------------------------------

let backend_url = $"http://($local_ip):($backend_port)"
let frontend_url = $"http://($local_ip):($frontend_port)"
mut actual_frontend_url = $frontend_url

if $mode == "all" or $mode == "backend" {
    print $"🔧 Bygger Go backend..."
    cd /home/simon/repos/begbot
    let build_result = (^bash -c "go build -o ./tmp/main ./cmd/api 2>&1" | complete)
    if $build_result.exit_code != 0 {
        print $"❌ Byggfel:\n($build_result.stdout)"
        exit 1
    }
    print "✓ Bygget klart"

    print $"🔧 Startar Go backend på port ($backend_port)..."
    if $have_socat == "yes" {
        let backend_proxy_cmd = $"socat TCP-LISTEN:($backend_port),reuseaddr,fork TCP:127.0.0.1:($backend_internal_port) >/dev/null 2>&1 &"
        ^bash -c $backend_proxy_cmd
        print $"✓ Socat proxy up: 127.0.0.1:($backend_port) -> 127.0.0.1:($backend_internal_port)"

        ^bash -c $"export PORT=($backend_internal_port) && ./tmp/main > /dev/null 2>&1 &"
        let backend_url = $"http://($local_ip):($backend_port)"
        print $"✓ Backend startad - intern port: ($backend_internal_port)"
        print $"✓ Backend publik URL: ($backend_url)"
    } else {
        ^bash -c $"export PORT=($backend_port) && ./tmp/main > /dev/null 2>&1 &"
        let backend_url = $"http://($local_ip):($backend_port)"
        print $"✓ Backend startad: ($backend_url)"
    }
}

if $mode == "all" or $mode == "frontend" {
    print $"🌐 Startar Nuxt frontend på port ($frontend_port)..."
    cd /home/simon/repos/begbot/frontend
    # Starta Nuxt på den valda porten och ge frontend information om backend-porten
    let log_file = "/tmp/nuxt-dev.log"
    if $have_socat == "yes" {
        # Bind publik frontend_port via socat till en intern port där Nuxt startas
        let frontend_proxy_cmd = $"socat TCP-LISTEN:($frontend_port),reuseaddr,fork TCP:127.0.0.1:($frontend_internal_port) >/dev/null 2>&1 &"
        ^bash -c $frontend_proxy_cmd
        print $"✓ Socat proxy up: 127.0.0.1:($frontend_port) -> 127.0.0.1:($frontend_internal_port)"

        # Starta Nuxt så den lyssnar på internal-port och pekar mot backend publik-port
        ^bash -c $"export PORT=($frontend_internal_port) HOST=127.0.0.1 API_BASE_URL='http://127.0.0.1:($backend_port)' && npm run dev > ($log_file) 2>&1 &"
        # Vänta lite så Nuxt hinner starta
        sleep 0.5sec
        $actual_frontend_url = $"http://($local_ip):($frontend_port)"
        print $"✓ Frontend startad - intern port: ($frontend_internal_port)"
        print $"✓ Frontend publik URL: ($actual_frontend_url)"
        try { ^bash -c $"xdg-open ($actual_frontend_url) >/dev/null 2>&1 &" } catch { }
    } else {
        # Fallback: starta direkt på publik port
        ^bash -c $"export PORT=($frontend_port) HOST=127.0.0.1 API_BASE_URL='http://127.0.0.1:($backend_port)' && npm run dev > ($log_file) 2>&1 &"
        sleep 0.5sec
        $actual_frontend_url = $"http://($local_ip):($frontend_port)"
        print $"✓ Frontend startad: ($actual_frontend_url)"
        try { ^bash -c $"xdg-open ($actual_frontend_url) >/dev/null 2>&1 &" } catch { }
    }
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
    curl -d $"Backend: ($backend_url) | Frontend: ($final_frontend_url) | .dev.env: ($devenv_file)" ntfy.sh/simonbrundin-dev-notification
} catch {
    print "Kunde inte skicka notifiering"
}

print "\n✅ Utvecklingsservrar körs!"
print $"   Backend: ($backend_url)"
print $"   Frontend: ($final_frontend_url)"
print $"   .dev.env: ($devenv_file)"
print $"\nFör att stoppa: kör om dev.nu, eller döda portarna manuellt: lsof -i :($backend_port) -t | xargs kill -9"
