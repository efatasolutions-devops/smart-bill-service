# Smart Bill Service - VPS Deployment Guide

## Prerequisites

### 1. VPS Requirements
- Ubuntu 20.04+ atau CentOS 7+
- RAM minimal 2GB
- Storage minimal 20GB
- Port 3000 terbuka (atau port yang Anda inginkan)

### 2. Software yang Diperlukan
- Go 1.19+
- Git
- Nginx (opsional, untuk reverse proxy)
- PM2 atau systemd (untuk process management)

## Step 1: Setup VPS Environment

### Install Go
```bash
# Download Go
wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz

# Extract
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### Install Git
```bash
sudo apt update
sudo apt install git -y
```

## Step 2: Clone Repository

```bash
# Clone repository
git clone https://github.com/efatasolutions-devops/smart-bill-service.git
cd smart-bill-service
```

## Step 3: Environment Configuration

### Create .env file
```bash
cp .env.example .env  # jika ada
# atau buat manual:
nano .env
```

### Sample .env Configuration
```env
# Server Configuration
PORT=3000
HOST=0.0.0.0

# Firebase Configuration
FIREBASE_SERVICE_ACCOUNT_KEY_PATH=./storage/splitbill-firebase-adminsdk.json
FIREBASE_PROJECT_ID=splitbill-4c851

# Storage Configuration
BUCKET_STORAGE=FIREBASE
# atau BUCKET_STORAGE=VM untuk local storage

# Google Gemini AI (jika digunakan)
GEMINI_API_KEY=your_gemini_api_key_here

# Logging
LOG_LEVEL=info
```

## Step 4: Firebase Setup

### Upload Firebase Service Account Key
```bash
# Buat direktori storage jika belum ada
mkdir -p storage

# Upload file splitbill-firebase-adminsdk.json ke storage/
# Gunakan scp, rsync, atau upload manual
scp splitbill-firebase-adminsdk.json user@your-vps:/path/to/smart-bill-service/storage/
```

## Step 5: Build and Run Application

### Install Dependencies
```bash
go mod tidy
go mod download
```

### Build Application
```bash
go build -o smart-bill-service main.go
```

### Test Run
```bash
./smart-bill-service
```

## Step 6: Production Deployment

### Option A: Using Systemd (Recommended)

#### Create systemd service file
```bash
sudo nano /etc/systemd/system/smart-bill-service.service
```

#### Service Configuration
```ini
[Unit]
Description=Smart Bill Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/smart-bill-service
ExecStart=/home/ubuntu/smart-bill-service/smart-bill-service
Restart=always
RestartSec=5
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
Environment=GOPATH=/home/ubuntu/go

[Install]
WantedBy=multi-user.target
```

#### Start Service
```bash
sudo systemctl daemon-reload
sudo systemctl enable smart-bill-service
sudo systemctl start smart-bill-service
sudo systemctl status smart-bill-service
```

### Option B: Using PM2

#### Install PM2
```bash
# Install Node.js dan npm
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install PM2
sudo npm install -g pm2
```

#### Create PM2 ecosystem file
```bash
nano ecosystem.config.js
```

```javascript
module.exports = {
  apps: [{
    name: 'smart-bill-service',
    script: './smart-bill-service',
    cwd: '/home/ubuntu/smart-bill-service',
    instances: 1,
    autorestart: true,
    watch: false,
    max_memory_restart: '1G',
    env: {
      NODE_ENV: 'production',
      PORT: 3000
    }
  }]
}
```

#### Start with PM2
```bash
pm2 start ecosystem.config.js
pm2 save
pm2 startup
```

## Step 7: Nginx Reverse Proxy (Optional)

### Install Nginx
```bash
sudo apt install nginx -y
```

### Configure Nginx
```bash
sudo nano /etc/nginx/sites-available/smart-bill-service
```

```nginx
server {
    listen 80;
    server_name your-domain.com;  # Ganti dengan domain Anda

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### Enable Site
```bash
sudo ln -s /etc/nginx/sites-available/smart-bill-service /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

## Step 8: Firewall Configuration

```bash
# Allow SSH
sudo ufw allow ssh

# Allow HTTP/HTTPS
sudo ufw allow 80
sudo ufw allow 443

# Allow application port (jika tidak menggunakan Nginx)
sudo ufw allow 3000

# Enable firewall
sudo ufw enable
```

## Step 9: SSL Certificate (Optional)

### Using Certbot for Let's Encrypt
```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d your-domain.com
```

## Monitoring and Maintenance

### Check Application Status
```bash
# Systemd
sudo systemctl status smart-bill-service

# PM2
pm2 status
pm2 logs smart-bill-service
```

### View Logs
```bash
# Application logs (jika menggunakan file logging)
tail -f storage/logs/general_log/$(date +%d-%m-%Y).log

# Systemd logs
sudo journalctl -u smart-bill-service -f
```

### Update Application
```bash
cd smart-bill-service
git pull origin main
go build -o smart-bill-service main.go

# Restart service
sudo systemctl restart smart-bill-service
# atau
pm2 restart smart-bill-service
```

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   sudo lsof -i :3000
   sudo kill -9 <PID>
   ```

2. **Permission denied for Firebase file**
   ```bash
   chmod 600 storage/splitbill-firebase-adminsdk.json
   ```

3. **Go module issues**
   ```bash
   go clean -modcache
   go mod tidy
   ```

4. **Storage directory permissions**
   ```bash
   mkdir -p storage/logs/general_log
   chmod -R 755 storage/
   ```

## Security Recommendations

1. **Firewall**: Hanya buka port yang diperlukan
2. **SSH**: Gunakan key-based authentication
3. **Updates**: Selalu update sistem secara berkala
4. **Backup**: Backup konfigurasi dan data secara rutin
5. **Monitoring**: Setup monitoring untuk aplikasi dan server

## API Testing

Setelah deployment, test API dengan:

```bash
# Health check
curl http://your-domain.com/

# Upload test (dengan file gambar)
curl -X POST -F "image=@test-receipt.jpg" http://your-domain.com/

# Swagger documentation
curl http://your-domain.com/swagger/index.html
