package utils

import (
	"log/slog"
	"os"
)

func HandleErrorOrLogWithMessages(err error, errMsg, successMsg string) {
	if err != nil {
		if errMsg == "" {
			errMsg = "operation failed"
		}
		slog.Error(errMsg, "error", err)
		os.Exit(1)
	}

	if successMsg != "" {
		slog.Info(successMsg)
	}
}
