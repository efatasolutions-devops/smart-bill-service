package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var GeneralLogger *logrus.Logger

func setupLogger(filename string) *logrus.Logger {
	// numberInt,_ := strconv.Atoi(os.Getenv("TIME_STORAGE_DAY"))
	permission_r := os.Getenv("LOG_PERMISSION_READ")

	permission_r_num, err := strconv.ParseUint(permission_r, 8, 32)
	if err != nil {
		permission_r_num = 0666 // Set default value
		log.Fatalf("Failed to open the log file with the specified permissions")
	}

	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(permission_r_num))
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	permission_chmod := os.Getenv("LOG_PERMISSION_CHMOD")

	permission_chmod_num, err := strconv.ParseUint(permission_chmod, 8, 32)
	if err != nil {
		permission_chmod_num = 0777 // Set default value
		log.Fatalf("Failed to set the log file with the specified permissions")
	}
	// Set file permissions to 777
	err = os.Chmod(filename, os.FileMode(permission_chmod_num))
	if err != nil {
		log.Fatalf("Failed to change file permissions: %v", err)
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    999999999999999999,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   false,
		LocalTime:  true, // Use local time for log rotation
	}

	log := logrus.New()
	// logrus.SetReportCaller(true)
	log.SetReportCaller(true)
	log.SetOutput(lumberjackLogger)
	log.SetFormatter(&logrus.JSONFormatter{
		DisableHTMLEscape: true,
	})
	return log
}

func Logger(app *fiber.App) {

	// date := time.Now().Format("01-02-2006")
	// generalLogFile := fmt.Sprintf("./storage/logs/general_log/%s.log", date)
	// dailyLogFile := fmt.Sprintf("./storage/logs/%s.log", date)

	// GeneralLogger = setupLogger(generalLogFile)
	// dailyLogger := setupLogger(dailyLogFile)
	createDirStorageLogs()

	app.Use(func(c *fiber.Ctx) error {
		// defer helpers.RecoverPanicContext(c)
		defer func() error {
			if r := recover(); r != nil {
				err := fmt.Sprintf("Error occured panic %s", r)
				return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
					"data":   nil,
					"status": err,
				})
			}
			return nil
		}()
		date := time.Now().Format("01-02-2006")
		generalLogFile := fmt.Sprintf("./storage/logs/general_log/%s.log", date)
		dailyLogFile := fmt.Sprintf("./storage/logs/%s.log", date)
		GeneralLogger = setupLogger(generalLogFile)
		dailyLogger := setupLogger(dailyLogFile)

		start := time.Now()
		err := c.Next()
		latency := time.Since(start)
		latencyStr := fmt.Sprintf("%dms", latency.Milliseconds())

		body := c.Body()
		var bodyJSON interface{}
		if err := json.Unmarshal(body, &bodyJSON); err == nil {
			compactBody, _ := json.Marshal(bodyJSON)
			c.Locals("body", string(compactBody))
		} else {
			c.Locals("body", string(body))
		}

		var unqouteBody map[string]interface{}
		_ = json.Unmarshal([]byte(c.Locals("body").(string)), &unqouteBody)

		var unqouteResBody map[string]interface{}
		_ = json.Unmarshal([]byte(string(c.Response().Body())), &unqouteResBody)

		// string(c.Response().Body())
		GeneralLogger.WithFields(logrus.Fields{
			"body":         unqouteBody,
			"queryParams":  c.OriginalURL(),
			"reqHeaders":   c.GetReqHeaders(),
			"time":         time.Now().Format("15:04:05"),
			"date":         date,
			"status":       c.Response().StatusCode(),
			"ip":           c.IP(),
			"method":       c.Method(),
			"url":          c.OriginalURL(),
			"path":         c.Path(),
			"route":        c.Route().Path,
			"error":        err,
			"resBody":      unqouteResBody,
			"responseTime": latencyStr,
		}).Info("Request logged")

		resBody := string(c.Response().Body())

		// dailyLogger.WithFields(logrus.Fields{
		// 	"body":          c.Locals("body"),
		// 	"queryParams":   c.OriginalURL(),
		// 	"reqHeaders":    c.GetReqHeaders(),
		// 	"time":          time.Now().Format("15:04:05"),
		// 	"date":          date,
		// 	"status":        c.Response().StatusCode(),
		// 	"ip":            c.IP(),
		// 	"method":        c.Method(),
		// 	"url":           c.OriginalURL(),
		// 	"path":          c.Path(),
		// 	"route":         c.Route().Path,
		// 	"error":         err,
		// 	"resBody":       string(c.Response().Body()),
		// 	"responseTime":  latencyStr,
		// }).Info("Request logged")
		customLogEntry := fmt.Sprintf(
			"body : %s | queryParams : %s | reqHeaders : %v | time : %s | date : %s | status : %d | ip : %s | method : %s | url : %s | path : %s | route : %s | error : %v | resBody : %s | responseTime : %s",
			c.Locals("body"),
			c.OriginalURL(),
			c.GetReqHeaders(),
			time.Now().Format("15:04:05"),
			date,
			c.Response().StatusCode(),
			c.IP(),
			c.Method(),
			c.OriginalURL(),
			c.Path(),
			c.Route().Path,
			err,
			resBody,
			latencyStr,
		)

		dailyLogger.Out.Write([]byte(customLogEntry + "\n"))

		return nil
	})
}

func createDirStorageLogs() {
	dirs := []string{
		"./storage/logs/general_log",
		"./storage/logs",
	}
	permission_chmod := os.Getenv("LOG_PERMISSION_CHMOD")

	permission_chmod_num, err := strconv.ParseUint(permission_chmod, 8, 32)
	if err != nil {
		permission_chmod_num = 0777 // Set default value
		log.Fatalf("Failed to set the log file with the specified permissions")
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, os.FileMode(permission_chmod_num))
			if err != nil {
				fmt.Println(dir, "can't create directory")
			}
			fmt.Println("success created directory", dir)
		} else {
			fmt.Println("The provided directory named", dir, "exists")
		}
	}
}
