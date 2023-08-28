package intrumentation

import (
	"github.com/eliecharra/ghoma/internal/intrumentation/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(config *config.Config) error {
	cfg := zap.NewProductionConfig()
	if config.IsDev() {
		cfg = zap.NewDevelopmentConfig()
		cfg.Encoding = "console"
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	level, err := zap.ParseAtomicLevel(config.LogLevel)
	if err != nil {
		return err
	}
	cfg.Level = level

	logger, err := cfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		return err
	}

	logger.Sugar()

	_ = zap.ReplaceGlobals(logger)
	return nil
}
