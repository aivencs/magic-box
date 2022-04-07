// 此包支持结构体数据校验
package validate

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/go-playground/validator"
)

// 使用枚举限定选择
type SupportType string

const (
	Validator SupportType = "validator"
)

// 定义全局配置对象
var validate Validate
var locker = new(sync.Mutex)
var once sync.Once

// 抽象接口
type Validate interface {
	Work(ctx context.Context, payload interface{}) (string, error)
}

// 设置全局对象
func SetValidate(v Validate) {
	locker.Lock()
	defer locker.Unlock()
	validate = v
}

// 获取全局对象
func GetValidate() Validate {
	locker.Lock()
	defer locker.Unlock()
	return validate
}

// 初始化时所用参数
type Option struct{}

// 初始化对象
func InitValidate(ctx context.Context, name SupportType, option Option) error {
	var c = validate
	once.Do(func() {
		c = ValidateFactory(ctx, name, option)
	})
	if c == nil {
		return errors.New("初始化失败")
	}
	SetValidate(c)
	return nil
}

// 抽象工厂
func ValidateFactory(ctx context.Context, name SupportType, option Option) Validate {
	switch name {
	case Validator:
		return NewValidator(ctx, option)
	default:
		return NewValidator(ctx, option)
	}
}

// 结构体
// 基于
type TheValidator struct {
	Kernel *validator.Validate
}

// 创建基于的对象
func NewValidator(ctx context.Context, option Option) Validate {
	v := validator.New()
	// 将label作为字段名称
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("label")
		return name
	})
	return &TheValidator{Kernel: v}
}

func (c *TheValidator) Work(ctx context.Context, payload interface{}) (string, error) {
	message := ""
	err := c.Kernel.Struct(payload)
	if err == nil {
		return "", err
	}
	for _, err := range err.(validator.ValidationErrors) {
		switch err.Tag() {
		case "required":
			message = fmt.Sprintf("%s为必填项", err.Field())
		case "min":
			message = fmt.Sprintf("%s的长度不应小于%v", err.Field(), err.Param())
		case "max":
			message = fmt.Sprintf("%s的长度不应超过%v", err.Field(), err.Param())
		case "ne":
			message = fmt.Sprintf("%s的值不应为%v", err.Field(), err.Value())
		case "len":
			message = fmt.Sprintf("%s的长度必须为%v", err.Field(), err.Param())
		case "eq":
			message = fmt.Sprintf("%s的值必须为%v", err.Field(), err.Param())
		case "oneof":
			message = fmt.Sprintf("%s的值必须在[%v]其中", err.Field(), err.Param())
		case "gt":
			message = fmt.Sprintf("%s的值必须大于%v", err.Field(), err.Param())
		case "gte":
			message = fmt.Sprintf("%s的值必须大于或等于%v", err.Field(), err.Param())
		case "lt":
			message = fmt.Sprintf("%s的值必须小于%v", err.Field(), err.Param())
		case "lte":
			message = fmt.Sprintf("%s的值必须小于或等于%v", err.Field(), err.Param())
		case "eqfield":
			message = fmt.Sprintf("%s的值必须与%v的值相等", err.Field(), err.Param())
		case "numeric":
			message = fmt.Sprintf("%s的值必须为数字", err.Field())
		case "email":
			message = fmt.Sprintf("%s的值必须符合邮箱格式", err.Field())
		case "url":
			message = fmt.Sprintf("%s的值必须符合网址格式", err.Field())
		case "ip":
			message = fmt.Sprintf("%s的内容必须符合IP格式", err.Field())
		case "contains":
			message = fmt.Sprintf("%s的值必须包含%v", err.Field(), err.Param())
		case "excludes":
			message = fmt.Sprintf("%s的值不可包含%v", err.Field(), err.Param())
		case "containsany":
			message = fmt.Sprintf("%s的值必须包含[%v]其中任意一个", err.Field(), err.Param())
		case "excludesall":
			message = fmt.Sprintf("%s的值不可包含[%v]其中任意一个", err.Field(), err.Param())
		case "startswith":
			message = fmt.Sprintf("%s的值必须以[%v]为开头", err.Field(), err.Param())
		case "endswith":
			message = fmt.Sprintf("%s的值必须以[%v]为结尾", err.Field(), err.Param())
		default:
			message = fmt.Sprintf("%s的值未通过校验", err.Field())
		}
	}
	return message, err
}

func Work(ctx context.Context, payload interface{}) (string, error) {
	v := GetValidate()
	return v.Work(ctx, payload)
}
