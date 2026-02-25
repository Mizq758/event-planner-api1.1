package logs

import (
	"os"

	"github.com/labstack/gommon/log"
)

func NewLogger(writeToFile bool) *log.Logger {
	// Настраиваем логгер для записи в файл
	logger := log.New("dict")
	if writeToFile {
		// Создаем файл для записи логов
		logFile, err := os.OpenFile("app.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}

		logger.SetOutput(logFile)
	}

	logger.SetLevel(log.INFO) // Уровень логирования: DEBUG, INFO, WARN, ERROR
	logger.SetHeader("${time_rfc3339} ${level} ${short_file}:${line} ${message}")

	// Пример логирования
	logger.Info("Application started")

	return logger
}
