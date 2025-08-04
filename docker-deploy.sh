#!/bin/bash

# Smart Bill Service - Docker Deployment Script
# Usage: ./docker-deploy.sh [start|stop|restart|logs|status]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.yml"
PROJECT_NAME="smart-bill-service"

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

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
}

# Check if required files exist
check_files() {
    local required_files=("Dockerfile" "docker-compose.yml" "nginx.conf")
    
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            log_error "Required file not found: $file"
            exit 1
        fi
    done
}

# Setup environment
setup_environment() {
    log_info "Setting up environment..."
    
    # Create .env if not exists
    if [ ! -f ".env" ]; then
        cat > .env << EOF
# Server Configuration
PORT=3000
HOST=0.0.0.0

# Firebase Configuration
FIREBASE_SERVICE_ACCOUNT_KEY_PATH=./storage/splitbill-firebase-adminsdk.json
FIREBASE_PROJECT_ID=splitbill-4c851

# Storage Configuration
BUCKET_STORAGE=FIREBASE

# Logging
LOG_LEVEL=info
EOF
        log_warning "Created default .env file. Please update with your actual values!"
    fi
    
    # Create storage directories
    mkdir -p storage/logs/general_log
    mkdir -p storage/public/images
    
    # Check Firebase service account key
    if [ ! -f "storage/splitbill-firebase-adminsdk.json" ]; then
        log_warning "Firebase service account key not found at storage/splitbill-firebase-adminsdk.json"
        log_warning "Please upload your Firebase service account key to this location"
    fi
    
    log_success "Environment setup completed"
}

# Start services
start_services() {
    log_info "Starting Smart Bill Service with Docker..."
    
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d --build
    
    log_success "Services started successfully"
    show_status
}

# Stop services
stop_services() {
    log_info "Stopping Smart Bill Service..."
    
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down
    
    log_success "Services stopped successfully"
}

# Restart services
restart_services() {
    log_info "Restarting Smart Bill Service..."
    
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME restart
    
    log_success "Services restarted successfully"
    show_status
}

# Show logs
show_logs() {
    local service=${1:-}
    
    if [ -n "$service" ]; then
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f $service
    else
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f
    fi
}

# Show status
show_status() {
    log_info "Service Status:"
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps
    
    echo ""
    log_info "Container Health:"
    docker ps --filter "name=${PROJECT_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    
    echo ""
    log_info "Service URLs:"
    echo "  Application: http://localhost:3000"
    echo "  Nginx Proxy: http://localhost:80"
    echo "  Health Check: http://localhost/health"
}

# Update services
update_services() {
    log_info "Updating Smart Bill Service..."
    
    # Pull latest code (if in git repo)
    if [ -d ".git" ]; then
        git pull origin main
    fi
    
    # Rebuild and restart
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d --build
    
    log_success "Services updated successfully"
    show_status
}

# Clean up
cleanup() {
    log_info "Cleaning up Docker resources..."
    
    # Stop and remove containers
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down -v
    
    # Remove unused images
    docker image prune -f
    
    log_success "Cleanup completed"
}

# Install Docker (Ubuntu/Debian)
install_docker() {
    log_info "Installing Docker..."
    
    # Update package index
    sudo apt-get update
    
    # Install prerequisites
    sudo apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release
    
    # Add Docker's official GPG key
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    
    # Set up stable repository
    echo \
        "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
        $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # Install Docker Engine
    sudo apt-get update
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io
    
    # Install Docker Compose
    sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    
    # Add user to docker group
    sudo usermod -aG docker $USER
    
    log_success "Docker installed successfully"
    log_warning "Please log out and log back in for group changes to take effect"
}

# Show help
show_help() {
    echo "Smart Bill Service - Docker Deployment Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start       Start all services"
    echo "  stop        Stop all services"
    echo "  restart     Restart all services"
    echo "  update      Update and restart services"
    echo "  logs        Show logs (add service name for specific service)"
    echo "  status      Show service status"
    echo "  cleanup     Clean up Docker resources"
    echo "  install     Install Docker and Docker Compose"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start"
    echo "  $0 logs smart-bill-service"
    echo "  $0 status"
}

# Main function
main() {
    local command=${1:-help}
    
    case $command in
        "start")
            check_docker
            check_files
            setup_environment
            start_services
            ;;
        "stop")
            check_docker
            stop_services
            ;;
        "restart")
            check_docker
            restart_services
            ;;
        "update")
            check_docker
            check_files
            update_services
            ;;
        "logs")
            check_docker
            show_logs $2
            ;;
        "status")
            check_docker
            show_status
            ;;
        "cleanup")
            check_docker
            cleanup
            ;;
        "install")
            install_docker
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# Run main function
main "$@"
