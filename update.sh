#!/bin/bash

# Smart Bill Service - Update Script
# Usage: ./update.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="smart-bill-service"
APP_DIR="/home/$(whoami)/$APP_NAME"
SERVICE_NAME="smart-bill-service"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Backup current version
backup_current() {
    log_info "Creating backup of current version..."
    
    if [ -f "$APP_DIR/$APP_NAME" ]; then
        cp "$APP_DIR/$APP_NAME" "$APP_DIR/${APP_NAME}.backup.$(date +%Y%m%d_%H%M%S)"
        log_success "Backup created"
    else
        log_warning "No existing binary found to backup"
    fi
}

# Update repository
update_repository() {
    log_info "Updating repository..."
    
    cd $APP_DIR
    git fetch origin
    git pull origin main
    
    log_success "Repository updated"
}

# Build new version
build_application() {
    log_info "Building new version..."
    
    cd $APP_DIR
    
    # Set Go environment
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    
    # Install/update dependencies
    go mod tidy
    go mod download
    
    # Build
    go build -o $APP_NAME main.go
    chmod +x $APP_NAME
    
    log_success "New version built successfully"
}

# Restart service
restart_service() {
    log_info "Restarting service..."
    
    sudo systemctl restart $SERVICE_NAME
    sleep 3
    
    if sudo systemctl is-active --quiet $SERVICE_NAME; then
        log_success "Service restarted successfully"
    else
        log_error "Service failed to start"
        sudo journalctl -u $SERVICE_NAME --no-pager -n 20
        exit 1
    fi
}

# Verify update
verify_update() {
    log_info "Verifying update..."
    
    sleep 5
    
    # Check if service is running
    if sudo systemctl is-active --quiet $SERVICE_NAME; then
        log_success "Service is running"
    else
        log_error "Service is not running"
        exit 1
    fi
    
    # Check if port is listening
    if netstat -tuln | grep -q ":3000 "; then
        log_success "Application is listening on port 3000"
    else
        log_error "Application is not listening on port 3000"
        exit 1
    fi
    
    log_success "Update verification completed"
}

# Show status
show_status() {
    echo ""
    echo "=================================="
    echo "   UPDATE COMPLETED"
    echo "=================================="
    echo ""
    echo "Service Status:"
    sudo systemctl status $SERVICE_NAME --no-pager -l
    echo ""
    echo "Recent Logs:"
    sudo journalctl -u $SERVICE_NAME --no-pager -n 10
    echo ""
}

# Rollback function
rollback() {
    log_warning "Rolling back to previous version..."
    
    BACKUP_FILE=$(ls -t $APP_DIR/${APP_NAME}.backup.* 2>/dev/null | head -n1)
    
    if [ -n "$BACKUP_FILE" ]; then
        cp "$BACKUP_FILE" "$APP_DIR/$APP_NAME"
        chmod +x "$APP_DIR/$APP_NAME"
        sudo systemctl restart $SERVICE_NAME
        log_success "Rollback completed"
    else
        log_error "No backup file found for rollback"
        exit 1
    fi
}

# Main update function
main() {
    log_info "Starting Smart Bill Service update..."
    
    if [ ! -d "$APP_DIR" ]; then
        log_error "Application directory not found: $APP_DIR"
        log_error "Please run deployment first"
        exit 1
    fi
    
    backup_current
    update_repository
    build_application
    restart_service
    verify_update
    show_status
    
    log_success "Update completed successfully!"
}

# Handle command line arguments
case "${1:-}" in
    "rollback")
        rollback
        ;;
    *)
        main
        ;;
esac
