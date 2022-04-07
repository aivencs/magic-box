package logger

import "unicode/utf8"

var erc map[MessageCode]ErrorCode
var defaultErrorCode = SUCCESS

type MessageCode uint

type ErrorCode struct {
	Name  string
	Code  MessageCode
	Level LoggerLevel
	Label string
}

const (
	SUCCESS     MessageCode = 10000 // 操作成功 info
	CHECK       MessageCode = 10001 // 请检查 info
	LIMITERROR  MessageCode = 10002 // 超限 error
	TIMEOUT     MessageCode = 10003 // 超时 error
	SUPWARN     MessageCode = 10004 // 补充数据 warn
	STATUSERROR MessageCode = 10005 // 非常规状态码 error
	EDERROR     MessageCode = 10006 // 编码或解码失败 error
	RPERROR     MessageCode = 10007 // 运行时参数错误 error
	PVERROR     MessageCode = 10008 // 参数未通过校验 error
	DVERROR     MessageCode = 10009 // 数据结果未通过校验 error
	RWARN       MessageCode = 10010 // 运行时发生异常 warn
	RPWARN      MessageCode = 10011 // 运行时发生错误 error
	CALLTIMEOUT MessageCode = 20001 // 调用超时 error
	CALLERROR   MessageCode = 20002 // 调用错误 error
	INTERRUPT   MessageCode = 30001 // 组件中断 fatal
)

func InitErrorCode() {
	erc = map[MessageCode]ErrorCode{
		SUCCESS:     {Code: SUCCESS, Level: INFO, Label: "操作成功"},
		CHECK:       {Code: CHECK, Level: INFO, Label: "请检查"},
		LIMITERROR:  {Code: LIMITERROR, Level: INFO, Label: "超限"},
		TIMEOUT:     {Code: TIMEOUT, Level: INFO, Label: "超时"},
		SUPWARN:     {Code: SUPWARN, Level: WARN, Label: "补充数据"},
		STATUSERROR: {Code: STATUSERROR, Level: ERROR, Label: "非常规状态码"},
		EDERROR:     {Code: EDERROR, Level: ERROR, Label: "编码或解码错误"},
		RPERROR:     {Code: RPERROR, Level: ERROR, Label: "运行时参数错误"},
		PVERROR:     {Code: PVERROR, Level: ERROR, Label: "参数未通过校验"},
		DVERROR:     {Code: DVERROR, Level: ERROR, Label: "数据结果未通过校验"},
		RWARN:       {Code: RWARN, Level: WARN, Label: "运行时发生异常"},
		RPWARN:      {Code: RPWARN, Level: WARN, Label: "运行时发生错误"},
		CALLTIMEOUT: {Code: CALLTIMEOUT, Level: ERROR, Label: "调用超时"},
		CALLERROR:   {Code: CALLERROR, Level: ERROR, Label: "调用错误"},
		INTERRUPT:   {Code: INTERRUPT, Level: FATAL, Label: "组件中断"},
	}
}

func GetDefaultErc() ErrorCode {
	value := erc[defaultErrorCode]
	return value
}

func GetErc(code MessageCode, label string) ErrorCode {
	value := erc[code]
	if utf8.RuneCountInString(label) > 1 {
		value.Label = label
	}
	return value
}
