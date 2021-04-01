package logger

type Level int8

// RFC5424 log message levels.
const (
	LevelAttack Level = iota //# under attack
	LevelCritical
	LevelAlert
	LevelEmergency
	LevelNone  //# logger is close
	LevelInfo  //
	LevelWarn  //
	LevelError //
	LevelDebug //# under debug mode
)

func (l Level) String() string {
	switch l {
	case LevelAttack:
		return "attack"
	case LevelCritical:
		return "critical"
	case LevelAlert:
		return "alert"
	case LevelEmergency:
		return "emergency"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelDebug:
		return "debug"
	}
	return ""
}
