#!/bin/bash

# Smart Bill Service - Auto Deployment Script
# Usage: ./deploy.sh [production|staging]

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
REPO_URL="https://github.com/efatasolutions-devops/smart-bill-service.git"
GO_VERSION="1.21.0"

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

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_error "This script should not be run as root"
        exit 1
    fi
}

# Install Go if not exists
install_go() {
    if ! command -v go &> /dev/null; then
        log_info "Installing Go $GO_VERSION..."
        
        cd /tmp
        wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz
        sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
        
        # Add to PATH
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        source ~/.bashrc
        
        export PATH=$PATH:/usr/local/go/bin
        export GOPATH=$HOME/go
        
        log_success "Go installed successfully"
    else
        log_info "Go is already installed: $(go version)"
    fi
}

# Install system dependencies
install_dependencies() {
    log_info "Installing system dependencies..."
    
    sudo apt update
    sudo apt install -y git curl wget nginx ufw
    
    log_success "System dependencies installed"
}

# Clone or update repository
setup_repository() {
    if [ -d "$APP_DIR" ]; then
        log_info "Updating existing repository..."
        cd $APP_DIR
        git pull origin main
    else
        log_info "Cloning repository..."
        git clone $REPO_URL $APP_DIR
        cd $APP_DIR
    fi
    
    log_success "Repository setup completed"
}

# Setup environment file
setup_environment() {
    log_info "Setting up environment configuration..."
    
    if [ ! -f "$APP_DIR/.env" ]; then
        cat > $APP_DIR/.env << EOF
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
    else
        log_info ".env file already exists"
    fi
}

# Setup storage directories
setup_storage() {
    log_info "Setting up storage directories..."
    
    mkdir -p $APP_DIR/storage/logs/general_log
    mkdir -p $APP_DIR/storage/public/images
    
    # Set permissions
    chmod -R 755 $APP_DIR/storage/
    
    log_success "Storage directories created"
}

# Build application
build_application() {
    log_info "Building application..."
    
    cd $APP_DIR
    
    # Set Go environment
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    
    # Install dependencies
    go mod tidy
    go mod download
    
    # Build
    go build -o $APP_NAME main.go
    
    # Make executable
    chmod +x $APP_NAME
    
    log_success "Application built successfully"
}

# Create systemd service
create_systemd_service() {
    log_info "Creating systemd service..."
    
    sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null << EOF
[Unit]
Description=Smart Bill Service
After=network.target

[Service]
Type=simple
User=$(whoami)
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/$APP_NAME
Restart=always
RestartSec=5
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
Environment=GOPATH=$HOME/go

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable $SERVICE_NAME
    
    log_success "Systemd service created"
}

# Setup Nginx
setup_nginx() {
    log_info "Setting up Nginx reverse proxy..."
    
    sudo tee /etc/nginx/sites-available/$APP_NAME > /dev/null << EOF
server {
    listen 80;
    server_name _;

    client_max_body_size 50M;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        # Timeout settings
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
EOF

    # Enable site
    sudo ln -sf /etc/nginx/sites-available/$APP_NAME /etc/nginx/sites-enabled/
    sudo rm -f /etc/nginx/sites-enabled/default
    
    # Test configuration
    sudo nginx -t
    
    log_success "Nginx configured"
}

# Setup firewall
setup_firewall() {
    log_info "Configuring firewall..."
    
    # Reset UFW to defaults
    sudo ufw --force reset
    
    # Set default policies
    sudo ufw default deny incoming
    sudo ufw default allow outgoing
    
    # Allow SSH
    sudo ufw allow ssh
    
    # Allow HTTP/HTTPS
    sudo ufw allow 80
    sudo ufw allow 443
    
    # Enable firewall
    sudo ufw --force enable
    
    log_success "Firewall configured"
}

# Start services
start_services() {
    log_info "Starting services..."
    
    # Start application
    sudo systemctl restart $SERVICE_NAME
    sudo systemctl status $SERVICE_NAME --no-pager
    
    # Start Nginx
    sudo systemctl restart nginx
    sudo systemctl status nginx --no-pager
    
    log_success "Services started"
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."
    
    sleep 5
    
    # Check if service is running
    if sudo systemctl is-active --quiet $SERVICE_NAME; then
        log_success "Application service is running"
    else
        log_error "Application service is not running"
        sudo journalctl -u $SERVICE_NAME --no-pager -n 20
        exit 1
    fi
    
    # Check if port is listening
    if netstat -tuln | grep -q ":3000 "; then
        log_success "Application is listening on port 3000"
    else
        log_error "Application is not listening on port 3000"
        exit 1
    fi
    
    # Test HTTP endpoint
    if curl -f -s http://localhost:80/health > /dev/null; then
        log_success "HTTP endpoint is responding"
    else
        log_warning "HTTP endpoint test failed (this might be normal if no health endpoint exists)"
    fi
    
    log_success "Deployment verification completed"
}

# Show deployment info
show_info() {
    echo ""
    echo "=================================="
    echo "   DEPLOYMENT COMPLETED"
    echo "=================================="
    echo ""
    echo "Application: $APP_NAME"
    echo "Directory: $APP_DIR"
    echo "Service: $SERVICE_NAME"
    echo ""
    echo "Useful commands:"
    echo "  sudo systemctl status $SERVICE_NAME"
    echo "  sudo systemctl restart $SERVICE_NAME"
    echo "  sudo journalctl -u $SERVICE_NAME -f"
    echo "  tail -f $APP_DIR/storage/logs/general_log/\$(date +%d-%m-%Y).log"
    echo ""
    echo "Next steps:"
    echo "1. Update .env file with your actual configuration"
    echo "2. Upload Firebase service account key to $APP_DIR/storage/"
    echo "3. Restart service: sudo systemctl restart $SERVICE_NAME"
    echo "4. Test API: curl http://your-server-ip/"
    echo ""
}

# Main deployment function
main() {
    log_info "Starting Smart Bill Service deployment..."
    
    check_root
    install_dependencies
    install_go
    setup_repository
    setup_environment
    setup_storage
    build_application
    create_systemd_service
    setup_nginx
    setup_firewall
    start_services
    verify_deployment
    show_info
    
    log_success "Deployment completed successfully!"
}

# Run main function
main "$@"
