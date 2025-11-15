# Running Your App Without Sudo (Hackclub Nest Environment)

Since you don't have `sudo` access, here are your best options:

## Option 1: Using `nohup` + `screen` or `tmux` (Simplest)

### Using `nohup` with background process:
```bash
cd /home/anubhav/code/gh-stats-gif-server
nohup ./ghapp > server.log 2>&1 &
```

This runs the app in the background and continues even after logout. Logs go to `server.log`.

**To check if it's running:**
```bash
ps aux | grep ghapp
```

**To stop it:**
```bash
pkill -f ghapp
```

---

## Option 2: Using `screen` (Recommended - Better Control)

`screen` allows you to create a detachable session that persists after logout.

### Start the app in a screen session:
```bash
cd /home/anubhav/code/gh-stats-gif-server
screen -S gh-stats -d -m ./ghapp
```

This creates a detached screen session named `gh-stats` running your app.

**List active sessions:**
```bash
screen -ls
```

**View logs/output (attach to session):**
```bash
screen -r gh-stats
```

**Detach from session (without killing it):**
Press `Ctrl+A` then `D`

**Stop the app:**
```bash
screen -S gh-stats -X quit
```

---

## Option 3: Using `tmux` (More Modern Alternative to screen)

If `tmux` is available, it's similar to `screen` but more powerful.

### Start the app in a tmux session:
```bash
cd /home/anubhav/code/gh-stats-gif-server
tmux new-session -d -s gh-stats "./ghapp"
```

**List active sessions:**
```bash
tmux list-sessions
```

**View the app output:**
```bash
tmux attach-session -t gh-stats
```

**Detach:**
Press `Ctrl+B` then `D`

**Stop the app:**
```bash
tmux kill-session -t gh-stats
```

---

## Option 4: Using a Caddyfile with Your App

If you want to use **Caddy** (which is often available in Hackclub nests), here's how:

### First, ensure your app is running with Option 1, 2, or 3

### Create a `Caddyfile`:
```caddy
localhost:80 {
    reverse_proxy localhost:8800
}
```

Or with your domain:
```caddy
yourdomain.com {
    reverse_proxy localhost:8800
}
```

### Start Caddy (non-root):
```bash
caddy run --config Caddyfile
```

This reverse-proxies traffic from port 80 (or your domain) to your app on port 8800.

**Keep Caddy running in background:**
```bash
nohup caddy run --config Caddyfile > caddy.log 2>&1 &
```

---

## Option 5: Create a Simple Startup Script

Create `start-server.sh`:

```bash
#!/bin/bash

# Start the GitHub stats server
cd /home/anubhav/code/gh-stats-gif-server
nohup ./ghapp > server.log 2>&1 &

# Optional: Start Caddy for reverse proxy
# cd /home/anubhav/code/gh-stats-gif-server
# nohup caddy run --config Caddyfile > caddy.log 2>&1 &

echo "GitHub Stats Server started!"
echo "Logs: tail -f /home/anubhav/code/gh-stats-gif-server/server.log"
```

Make it executable:
```bash
chmod +x start-server.sh
```

Start it:
```bash
./start-server.sh
```

---

## Option 6: Using a Process Manager (If Available)

Some hackclub nests might have `supervisor` or similar. Check:
```bash
which supervisord
```

If available, ask your nest admin to add your app, or create a user config.

---

## Recommended Setup for Hackclub Nest

**Best combination:**

1. **Build your app:**
   ```bash
   cd /home/anubhav/code/gh-stats-gif-server
   go build -o ghapp main.go
   ```

2. **Start with screen:**
   ```bash
   screen -S gh-stats -d -m ./ghapp
   ```

3. **Test it's running:**
   ```bash
   curl http://localhost:8800/?id=octocat
   ```

4. **Check anytime:**
   ```bash
   screen -ls
   screen -r gh-stats
   ```

5. **Stop when needed:**
   ```bash
   screen -S gh-stats -X quit
   ```

---

## Caveats Without Sudo

⚠️ **Important limitations:**

- ❌ Can't use ports below 1024 (1-1023) without sudo - your port 8800 is fine
- ❌ App will stop if your user session is terminated by admin
- ❌ No auto-restart on server reboot
- ✅ Will survive your SSH logout
- ✅ Can manage with screen/tmux/nohup
- ✅ Can use Caddy for reverse proxying (if available)

---

## My Recommendation

**Use Option 2 (screen)** - it's:
- ✅ Simple
- ✅ Persistent
- ✅ Easy to manage
- ✅ No extra dependencies usually needed
- ✅ Can attach/detach easily
- ✅ Works everywhere

```bash
screen -S gh-stats -d -m ./ghapp
```

That's it! Your app runs in the background and survives SSH logout.

