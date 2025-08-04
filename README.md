# Smart Bill Service

Smart Bill Service adalah aplikasi Go yang menggunakan OCR dan AI untuk mengekstrak informasi dari gambar struk belanja. Aplikasi ini dapat mengidentifikasi item, harga, informasi toko, dan detail transaksi dari foto struk.

## 🚀 Fitur

- **OCR Processing**: Ekstraksi teks dari gambar struk menggunakan teknologi OCR
- **AI Analysis**: Analisis cerdas menggunakan Google Gemini AI untuk parsing data terstruktur
- **Firebase Storage**: Penyimpanan gambar di Firebase Storage atau lokal
- **RESTful API**: API endpoint yang mudah digunakan
- **Swagger Documentation**: Dokumentasi API interaktif
- **Docker Support**: Deployment menggunakan Docker dan Docker Compose
- **Nginx Reverse Proxy**: Load balancing dan SSL termination

## 📋 Requirements

### Sistem Requirements
- Go 1.19+ (untuk development)
- Docker & Docker Compose (untuk deployment)
- Ubuntu 20.04+ atau CentOS 7+ (untuk VPS deployment)
- RAM minimal 2GB
- Storage minimal 20GB

### API Keys & Credentials
- Firebase Service Account Key
- Google Gemini AI API Key (opsional)

## 🛠️ Installation & Deployment

### Option 1: Docker Deployment (Recommended)

#### Quick Start
```bash
# Clone repository
git clone https://github.com/efatasolutions-devops/smart-bill-service.git
cd smart-bill-service

# Make scripts executable
chmod +x docker-deploy.sh

# Start services
./docker-deploy.sh start
```

#### Docker Commands
```bash
# Start all services
./docker-deploy.sh start

# Stop services
./docker-deploy.sh stop

# Restart services
./docker-deploy.sh restart

# View logs
./docker-deploy.sh logs

# Check status
./docker-deploy.sh status

# Update services
./docker-deploy.sh update

# Clean up
./docker-deploy.sh cleanup
```

### Option 2: VPS Deployment

#### Automated Deployment
```bash
# Clone repository
git clone https://github.com/efatasolutions-devops/smart-bill-service.git
cd smart-bill-service

# Make scripts executable
chmod +x deploy.sh update.sh

# Run deployment
./deploy.sh
```

#### Manual Deployment
Lihat [DEPLOYMENT.md](DEPLOYMENT.md) untuk panduan deployment manual yang lengkap.

## ⚙️ Configuration

### Environment Variables

Buat file `.env` di root directory:

```env
# Server Configuration
PORT=3000
HOST=0.0.0.0

# Firebase Configuration
FIREBASE_SERVICE_ACCOUNT_KEY_PATH=./storage/splitbill-firebase-adminsdk.json
FIREBASE_PROJECT_ID=splitbill-4c851

# Storage Configuration
BUCKET_STORAGE=FIREBASE  # atau VM untuk local storage

# Google Gemini AI
GEMINI_API_KEY=your_gemini_api_key_here

# Logging
LOG_LEVEL=info
```

### Firebase Setup

1. Download Firebase Service Account Key dari Firebase Console
2. Simpan file sebagai `storage/splitbill-firebase-adminsdk.json`
3. Pastikan Firebase Storage sudah diaktifkan di project Anda

## 📚 API Documentation

### Endpoints

#### Upload Receipt
```http
POST /
Content-Type: multipart/form-data

Parameters:
- image: file (jpg, jpeg, png)
```

#### Response Format
```json
{
  "items": [
    {
      "name": "Nasi Goreng",
      "price": "25000.00",
      "quantity": "2",
      "total": "50000.00"
    }
  ],
  "store_information": {
    "store_name": "Restaurant ABC",
    "address": "Jl. Sudirman No. 123, Jakarta",
    "phone_number": "+62812345678",
    "email": "info@restaurant.com",
    "npwp": "12.345.678.9-012.345"
  },
  "totals": {
    "subtotal": "95000.00",
    "tax": {
      "name": "PPN",
      "amount": "5000.00",
      "dpp": "95000.00",
      "total_tax": "5000.00",
      "service_charge": "0.00"
    },
    "discount": "0.00",
    "total": "100000.00",
    "payment": "105000.00",
    "change": "5000.00"
  },
  "transaction_information": {
    "transaction_id": "TXN123456789",
    "date": "02/08/2025",
    "time": "19:30"
  }
}
```

### Swagger Documentation

Akses dokumentasi interaktif di: `http://your-server/swagger/index.html`

## 🧪 Testing

### Manual Testing
```bash
# Test dengan curl
curl -X POST -F "image=@test-receipt.jpg" http://localhost:3000/

# Test health endpoint
curl http://localhost/health
```

### Load Testing
```bash
# Install Apache Bench
sudo apt install apache2-utils

# Run load test
ab -n 100 -c 10 -p test-image.jpg -T multipart/form-data http://localhost/
```

## 📊 Monitoring & Logs

### View Logs

#### Docker Deployment
```bash
# All services
./docker-deploy.sh logs

# Specific service
./docker-deploy.sh logs smart-bill-service
```

#### VPS Deployment
```bash
# Application logs
sudo journalctl -u smart-bill-service -f

# Application file logs
tail -f storage/logs/general_log/$(date +%d-%m-%Y).log

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### Service Status
```bash
# Docker
./docker-deploy.sh status

# VPS
sudo systemctl status smart-bill-service
```

## 🔧 Maintenance

### Update Application

#### Docker
```bash
./docker-deploy.sh update
```

#### VPS
```bash
./update.sh
```

### Rollback (VPS only)
```bash
./update.sh rollback
```

### Backup
```bash
# Backup storage directory
tar -czf backup-$(date +%Y%m%d).tar.gz storage/

# Backup database (jika ada)
# mysqldump atau pg_dump commands
```

## 🔒 Security

### Firewall Configuration
```bash
# Allow HTTP/HTTPS
sudo ufw allow 80
sudo ufw allow 443

# Allow SSH
sudo ufw allow ssh

# Enable firewall
sudo ufw enable
```

### SSL Certificate
```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d your-domain.com
```

## 🐛 Troubleshooting

### Common Issues

#### Port Already in Use
```bash
sudo lsof -i :3000
sudo kill -9 <PID>
```

#### Firebase Permission Error
```bash
chmod 600 storage/splitbill-firebase-adminsdk.json
```

#### Go Module Issues
```bash
go clean -modcache
go mod tidy
```

#### Docker Issues
```bash
# Clean up Docker
docker system prune -a

# Rebuild containers
./docker-deploy.sh cleanup
./docker-deploy.sh start
```

### Debug Mode

Enable debug logging:
```env
LOG_LEVEL=debug
```

## 📁 Project Structure

```
smart-bill-service/
├── main.go                 # Entry point
├── go.mod                  # Go modules
├── go.sum                  # Go dependencies
├── .env                    # Environment variables
├── Dockerfile              # Docker configuration
├── docker-compose.yml      # Docker Compose configuration
├── nginx.conf              # Nginx configuration
├── deploy.sh               # VPS deployment script
├── update.sh               # Update script
├── docker-deploy.sh        # Docker deployment script
├── DEPLOYMENT.md           # Detailed deployment guide
├── config/                 # Configuration files
│   ├── database.go         # Database & Firebase config
│   ├── logger.go           # Logging configuration
│   └── appConfig/          # Application configuration
├── controllers/            # HTTP controllers
├── services/               # Business logic
├── helpers/                # Helper functions
├── models/                 # Data models
├── routes/                 # Route definitions
├── storage/                # Storage directory
│   ├── logs/               # Application logs
│   ├── public/             # Public files
│   └── splitbill-firebase-adminsdk.json  # Firebase key
└── docs/                   # Swagger documentation
```

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

- **Documentation**: [DEPLOYMENT.md](DEPLOYMENT.md)
- **Issues**: [GitHub Issues](https://github.com/efatasolutions-devops/smart-bill-service/issues)
- **Email**: support@efatasolutions.com

## 🚀 Quick Commands Reference

```bash
# Development
go run main.go
go build -o smart-bill-service main.go

# Docker Deployment
./docker-deploy.sh start
./docker-deploy.sh logs
./docker-deploy.sh status

# VPS Deployment
./deploy.sh
./update.sh
sudo systemctl status smart-bill-service

# Testing
curl -X POST -F "image=@test.jpg" http://localhost:3000/
curl http://localhost/health
