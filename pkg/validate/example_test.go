package validate

import (
	"context"
	"fmt"
	"log"
)

type Users struct {
	Phone   string `form:"phone" json:"phone" label:"手机号" validate:"required,max=12,min=11"`
	Passwd  string `form:"passwd" json:"passwd" label:"密码" validate:"required,max=20,min=6"`
	Code    string `form:"code" json:"code" label:"验证码" validate:"required,len=6"`
	Text    string `json:"text" label:"文本" validate:"oneof=red green"`
	Id      string `json:"id" label:"编号" validate:"required,numeric"`
	Confirm string `json:"confirm" label:"校验密码" validate:"eqfield=Passwd"`
	Email   string `json:"email" label:"邮箱" validate:"email"`
	Content string `json:"content" label:"正文" validate:"html"`
}

func ExampleValidator() {
	users := &Users{
		Phone:   "1092872222",
		Passwd:  "123098",
		Code:    "123456",
		Text:    "red",
		Confirm: "123098",
		Id:      "12",
		Email:   "abcfoxmail@foxmail.com",
		Content: "<a>aps<a>",
	}
	ctx := context.WithValue(context.Background(), "trace", "v001")
	err := InitValidate(ctx, "validator", Option{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(validate.Work(ctx, users))
}
