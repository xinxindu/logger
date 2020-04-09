# logger
## 日志库
```
// 初始化日志对象
func InitLogger(when string, backupCount int, level Level, fileDir, fileName string) (*Logger, error)

// 打印info日志
func ( l *Logger) Infof (format string, args ...interface{})

// 关闭日志对象
func ( l *Logger) Close() 
```


