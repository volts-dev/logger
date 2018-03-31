package logger

// RFC5424 log message levels.
const (
	LevelAttack = iota //# under attack
	LevelCritical
	LevelAlert
	LevelEmergency
	LevelNone  //# logger is close
	LevelInfo  //
	LevelWarn  //
	LevelError //
	LevelDebug //# under debug mode
)
