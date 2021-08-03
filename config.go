package logger

type (
	Option func(*TConfig)

	TConfig struct {
		Level  Level  `json:"Level"`
		Prefix string `json:"Prefix"`
	}
)

func WithPrefix(name string) Option {
	return func(cfg *TConfig) {
		cfg.Prefix = name
	}
}

func WithLevel(lvl Level) Option {
	return func(cfg *TConfig) {
		cfg.Level = lvl
	}
}
