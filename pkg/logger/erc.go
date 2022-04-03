package logger

type MessageCode uint

const (
	SUCCESS     MessageCode = 10000 // 操作成功 info
	CHECK       MessageCode = 10001 // 请检查 check
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
	CALLTIMEOUT MessageCode = 20001 // 调用超时 check
	CALLERROR   MessageCode = 20002 // 调用错误 error
	INTERRUPT   MessageCode = 30001 // 组件中断 fatal
)
