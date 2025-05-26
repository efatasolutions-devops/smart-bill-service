package config

import (
	"context"
	"fmt"
	"log"
	"os"

	// Pastikan Anda mengimpor kedua package storage ini
	// cloud.google.com/go/storage untuk BucketHandle
	"cloud.google.com/go/storage"
	// firebase.google.com/go/storage untuk Client yang dikembalikan Firebase SDK
	firebase "firebase.google.com/go"
	firebaseStorage "firebase.google.com/go/storage" // <--- ALIAS untuk menghindari konflik nama
	"google.golang.org/api/option"
	"gorm.io/gorm" // Hapus jika Anda sudah yakin tidak ada GORM yang dipakai
)

var DB *gorm.DB // Hapus ini jika tidak dipakai

var FirebaseApp *firebase.App

// Ubah deklarasi ini. Kini kita akan menyimpan *cloud.google.com/go/storage.BucketHandle
var FirebaseStorageBucket *storage.BucketHandle

func ConnectFirebase() {
	var err error

	// Inisialisasi Firebase
	serviceAccountKeyPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY_PATH")
	if serviceAccountKeyPath == "" {
		serviceAccountKeyPath = "./storage/splitbill-firebase-adminsdk.json" // Default path
	}

	opt := option.WithCredentialsFile(serviceAccountKeyPath)
	FirebaseApp, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v\n", err)
	}

	// Dapatkan client Firebase Storage
	var storageClient *firebaseStorage.Client // Gunakan alias yang kita buat
	storageClient, err = FirebaseApp.Storage(context.Background())
	if err != nil {
		log.Fatalf("Error getting Firebase Storage client: %v\n", err)
	}

	// Dapatkan handler untuk Firebase Storage bucket
	// Ini adalah bagian yang menyebabkan error, dan ini perbaikannya:
	// Panggil DefaultBucket() dari storageClient
	bucketName := "splitbill-4c851.firebasestorage.app"
	FirebaseStorageBucket, err = storageClient.Bucket(bucketName)
	if err != nil {
		log.Fatalf("Error getting default Firebase Storage bucket: %v\n", err)
	}

	// Opsional: Jika Anda ingin menentukan nama bucket secara eksplisit (ganti dengan ID proyek Firebase Anda)
	// Misalnya, jika nama bucket Anda adalah "my-project-12345.appspot.com"
	// Atau jika Anda membuat bucket terpisah di GCP
	// bucketName := "your-project-id.appspot.com"
	// FirebaseStorageBucket = storageClient.Bucket(bucketName)

	fmt.Println("Firebase initialized successfully!")

	// Hapus atau komentar bagian GORM/MySQL jika tidak digunakan
	// func PostgresSQL() *gorm.DB {} juga bisa dihapus jika tidak digunakan
}

// Hapus fungsi ini jika tidak digunakan sama sekali
func PostgresSQL() *gorm.DB {
	return nil // Mengembalikan nil karena tidak ada koneksi
}
