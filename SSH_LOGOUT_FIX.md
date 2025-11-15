# SSH Logout Issues: Why It Happens & Solutions

## Why Your Server Stops When You Logout

Even with `screen` or `nohup`, SSH can terminate background processes due to:

1. **HUP Signal (SIGHUP)** - SSH sends this when you disconnect
2. **Shell Configuration** - `.bashrc` or `.bash_profile` might have settings killing processes
3. **Nest Admin Settings** - Some hackclub nests configure SSH to kill all user processes on logout
4. **Terminal Multiplexer Issues** - `screen` might not be properly detached

---

## Solution: Use the Provided `run-server.sh` Script

This script uses **`nohup`** with proper signal handling to survive SSH logout.

### Step 1: Build Your App

```bash
cd /home/anubhav/code/gh-stats-gif-server
go build -o ghapp main.go
```

### Step 2: Start the Server

```bash
./run-server.sh start
```

**Output should look like:**
```
Starting GitHub Stats Server...
✓ Server started successfully (PID: 12345)
✓ Logs: tail -f /home/anubhav/code/gh-stats-gif-server/server.log
```

### Step 3: Test It Works

Before logging out, test the server:

```bash
curl http://localhost:8800/?id=octocat
```

### Step 4: Check Status (Even After SSH Logout)

```bash
./run-server.sh status
```

Output:
```
✓ Server is running (PID: 12345)
Command: /home/anubhav/code/gh-stats-gif-server/ghapp
```

### Step 5: View Logs Anytime

```bash
./run-server.sh logs
```

---

## Available Commands

```bash
./run-server.sh start        # Start the server
./run-server.sh stop         # Stop the server
./run-server.sh restart      # Restart the server
./run-server.sh status       # Check if running
./run-server.sh logs         # View live logs (tail -f)
```

---

## Why This Works (Technical Details)

The script uses:

1. **`nohup`** - Ignores SIGHUP signal when SSH disconnects
2. **`&`** - Runs in background (detaches from terminal)
3. **Output Redirection** - Sends logs to a file instead of terminal
4. **PID Tracking** - Stores process ID to verify status later

**Before (fails on SSH logout):**
```bash
./ghapp  # Dies when you logout
```

**After (survives SSH logout):**
```bash
nohup ./ghapp > server.log 2>&1 &  # Lives after logout
```

---

## Testing Process (Verification Steps)

### 1. Start the server:
```bash
./run-server.sh start
```

### 2. Verify it's running:
```bash
./run-server.sh status
```

### 3. Test the API:
```bash
curl http://localhost:8800/?id=octocat
```

### 4. **Logout SSH completely:**
```bash
exit
```

### 5. **SSH back in** and verify server still running:
```bash
./run-server.sh status
```

**If it says "✓ Server is running", you're good!**

---

## Advanced: Auto-Start on Login

If you want the server to automatically start when you SSH in, add this to your `~/.bashrc`:

```bash
# Auto-start GitHub stats server if not already running
if ! /home/anubhav/code/gh-stats-gif-server/run-server.sh status >/dev/null 2>&1; then
    /home/anubhav/code/gh-stats-gif-server/run-server.sh start >/dev/null 2>&1
fi
```

Then:
```bash
source ~/.bashrc
```

Now every time you SSH in, it checks and starts the server if needed.

---

## Troubleshooting

### Issue: "Server stops even with nohup"

**Check if your shell is killing processes:**
```bash
shopt | grep huponexit
```

If it says `huponexit on`, add this to `~/.bash_profile`:
```bash
shopt -s huponexit  # Disable killing processes on exit
```

Wait, that enables it. Actually do:
```bash
shopt -u huponexit  # Disable (no HUP on exit)
```

### Issue: "Can't find ghapp binary"

```bash
cd /home/anubhav/code/gh-stats-gif-server
go build -o ghapp main.go
ls -la ghapp  # Verify it exists
```

### Issue: "PID file error"

```bash
rm /home/anubhav/code/gh-stats-gif-server/server.pid
./run-server.sh start
```

### Issue: "Port 8800 already in use"

```bash
lsof -i :8800  # See what's using it
kill -9 <PID>  # Kill the process
./run-server.sh restart
```

---

## Complete Workflow Example

```bash
# SSH into server
ssh user@server

# Navigate to project
cd /home/anubhav/code/gh-stats-gif-server

# Build if needed
go build -o ghapp main.go

# Start the server
./run-server.sh start

# Verify it's running
./run-server.sh status

# Test it
curl http://localhost:8800/?id=octocat

# View logs if needed
./run-server.sh logs

# Logout - server KEEPS RUNNING
exit

# SSH back in later
ssh user@server

# Check server status - it's still running!
cd /home/anubhav/code/gh-stats-gif-server
./run-server.sh status
```

---

## How to Properly Shutdown

When you're done and want to stop the server:

```bash
./run-server.sh stop
```

Or stop it without checking status:
```bash
kill $(cat /home/anubhav/code/gh-stats-gif-server/server.pid)
```

---

## Next Steps

1. **Run:** `./run-server.sh start`
2. **Verify:** `./run-server.sh status`
3. **Test:** `curl http://localhost:8800/?id=octocat`
4. **Logout:** `exit`
5. **SSH back in** and run `./run-server.sh status` again

If it's still running, you're all set! 🎉

