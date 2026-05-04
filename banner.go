package main

import (
	"log/slog"
)

func printBanner(logger *slog.Logger) {
	logger.Info(`┳┓     ┓•        ┏┓┏┓      ┳┳┓┏┓┏┓`)
    logger.Info(`┣┫┏┓┏┓┏┫┓┏┏┓┏┓╋  ┣ ┗┓  ━━  ┃┃┃┃ ┃┃`)
    logger.Info(`┻┛┗┻┛┗┗┻┗┗┗┛┗┛┗  ┻ ┗┛      ┛ ┗┗┛┣┛`)
}
