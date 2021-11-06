package log

type Logger interface {
	Debug(msg string, extras ...map[string]interface{})
	Info(msg string, extras ...map[string]interface{})
	Warn(msg string, extras ...map[string]interface{})
	Error(msg string, extras ...map[string]interface{})
	Panic(msg string, extras ...map[string]interface{})
}
