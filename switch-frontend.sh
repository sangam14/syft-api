#!/bin/bash

# Script to switch between the original HTML frontend and the Vue.js frontend

# Function to display usage information
show_usage() {
  echo "Usage: $0 [html|vue]"
  echo "  html - Switch to the original HTML frontend"
  echo "  vue  - Switch to the Vue.js frontend (requires building first)"
  echo ""
  echo "Example: $0 vue"
}

# Check if an argument was provided
if [ $# -ne 1 ]; then
  show_usage
  exit 1
fi

# Get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Create backup directory if it doesn't exist
BACKUP_DIR="$DIR/static/frontend_backups"
mkdir -p "$BACKUP_DIR"

# Function to backup current frontend files
backup_current_frontend() {
  local timestamp=$(date +%Y%m%d%H%M%S)
  local backup_path="$BACKUP_DIR/backup_$timestamp"
  
  echo "Creating backup of current frontend at $backup_path"
  mkdir -p "$backup_path"
  
  # Copy current frontend files
  cp -r "$DIR/static/index.html" "$DIR/static/script.js" "$DIR/static/style.css" "$backup_path/" 2>/dev/null || true
  
  echo "Backup created at $backup_path"
}

# Switch to the original HTML frontend
switch_to_html() {
  echo "Switching to original HTML frontend..."
  
  # Check if backup exists
  if [ -d "$BACKUP_DIR" ] && [ "$(ls -A $BACKUP_DIR)" ]; then
    # Find the most recent backup
    local latest_backup=$(find "$BACKUP_DIR" -type d -name "backup_*" | sort -r | head -n 1)
    
    if [ -n "$latest_backup" ]; then
      echo "Restoring from backup: $latest_backup"
      cp -r "$latest_backup"/* "$DIR/static/"
      echo "Original HTML frontend restored successfully."
      return 0
    fi
  fi
  
  echo "No backup found. Cannot restore original HTML frontend."
  return 1
}

# Switch to the Vue.js frontend
switch_to_vue() {
  echo "Switching to Vue.js frontend..."
  
  # Check if Vue.js frontend is built
  if [ ! -d "$DIR/static/vue-frontend/dist" ]; then
    echo "Vue.js frontend is not built yet. Building now..."
    
    # Navigate to Vue.js frontend directory
    cd "$DIR/static/vue-frontend"
    
    # Check if node_modules exists
    if [ ! -d "node_modules" ]; then
      echo "Installing dependencies..."
      npm install
    fi
    
    # Build Vue.js frontend
    echo "Building Vue.js frontend..."
    npm run build
    
    if [ ! -d "dist" ]; then
      echo "Build failed. Vue.js frontend could not be built."
      return 1
    fi
  fi
  
  # Backup current frontend
  backup_current_frontend
  
  # Copy Vue.js frontend files to static directory
  echo "Copying Vue.js frontend files to static directory..."
  cp -r "$DIR/static/vue-frontend/dist"/* "$DIR/static/"
  
  echo "Vue.js frontend switched successfully."
  return 0
}

# Main logic
case "$1" in
  "html")
    switch_to_html
    ;;
  "vue")
    switch_to_vue
    ;;
  *)
    echo "Invalid option: $1"
    show_usage
    exit 1
    ;;
esac

echo "Done. Restart the server to see the changes."
