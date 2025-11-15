#!/bin/bash

# GitHub Stats Server Wrapper Script
# This script handles proper background execution and prevents SSH termination

APP_DIR="/home/anubhav/code/gh-stats-gif-server"
APP_NAME="ghapp"
APP_PATH="$APP_DIR/$APP_NAME"
LOG_FILE="$APP_DIR/server.log"
PID_FILE="$APP_DIR/server.pid"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to start the server
start_server() {
    if [ -f "$PID_FILE" ]; then
        OLD_PID=$(cat "$PID_FILE")
        if ps -p "$OLD_PID" > /dev/null 2>&1; then
            echo -e "${YELLOW}Server is already running (PID: $OLD_PID)${NC}"
            return 1
        fi
    fi

    echo -e "${GREEN}Starting GitHub Stats Server...${NC}"
    
    # Use nohup with & to ensure it survives SSH logout
    # Redirect both stdout and stderr to log file
    nohup "$APP_PATH" > "$LOG_FILE" 2>&1 &
    
    NEW_PID=$!
    echo "$NEW_PID" > "$PID_FILE"
    
    sleep 1
    
    if ps -p "$NEW_PID" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Server started successfully (PID: $NEW_PID)${NC}"
        echo -e "${GREEN}✓ Logs: tail -f $LOG_FILE${NC}"
        return 0
    else
        echo -e "${RED}✗ Failed to start server${NC}"
        return 1
    fi
}

# Function to stop the server
stop_server() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${YELLOW}Stopping server (PID: $PID)...${NC}"
            kill "$PID"
            sleep 1
            if ps -p "$PID" > /dev/null 2>&1; then
                kill -9 "$PID"
            fi
            rm "$PID_FILE"
            echo -e "${GREEN}✓ Server stopped${NC}"
            return 0
        else
            echo -e "${YELLOW}Server not running${NC}"
            rm "$PID_FILE"
            return 0
        fi
    else
        echo -e "${YELLOW}No PID file found${NC}"
        return 0
    fi
}

# Function to check status
status_server() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Server is running (PID: $PID)${NC}"
            echo -e "${GREEN}Command: $(ps -p $PID -o cmd=)${NC}"
            return 0
        else
            echo -e "${RED}✗ Server is not running (stale PID file)${NC}"
            rm "$PID_FILE"
            return 1
        fi
    else
        echo -e "${RED}✗ Server is not running${NC}"
        return 1
    fi
}

# Function to view logs
view_logs() {
    if [ -f "$LOG_FILE" ]; then
        echo -e "${GREEN}Following logs (Ctrl+C to exit)...${NC}"
        tail -f "$LOG_FILE"
    else
        echo -e "${YELLOW}No log file found yet${NC}"
    fi
}

# Main script logic
case "${1:-}" in
    start)
        start_server
        ;;
    stop)
        stop_server
        ;;
    restart)
        stop_server
        sleep 1
        start_server
        ;;
    status)
        status_server
        ;;
    logs)
        view_logs
        ;;
    *)
        echo "GitHub Stats Server Manager"
        echo ""
        echo "Usage: $0 {start|stop|restart|status|logs}"
        echo ""
        echo "Commands:"
        echo "  start   - Start the server in background"
        echo "  stop    - Stop the server"
        echo "  restart - Restart the server"
        echo "  status  - Check if server is running"
        echo "  logs    - View live logs (tail -f)"
        echo ""
        echo "Examples:"
        echo "  $0 start"
        echo "  $0 status"
        echo "  $0 logs"
        echo "  $0 stop"
        ;;
esac
