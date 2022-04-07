package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aivencs/magic-box/pkg/config"
)

type BindConf struct {
	Pre     BindConfPre
	Runtime BindRuntime
}

type BindConfPre struct {
	Name string `json:"name"`
}

type BindRuntime struct {
	Name string `json:"name"`
}

func main() {
	ctx := context.Background()
	bindConf := BindConf{} // 选择符合配置格式的结构体
	// 初始化配置对象
	err := config.InitConf(ctx, config.Consul, config.Option{
		Application: "spanic-test",
		Env:         "dev",
		Auth:        false,
		Type:        "yaml",
		Bind:        &bindConf,
		Update:      true,
		Interval:    10,
	})
	if err != nil {
		log.Fatal(err)
	}
	// 使用方法
	for i := 0; i < 1000; i++ {
		fmt.Println("bind-", i, ": ", bindConf.Runtime.Name) // 直接访问
		// 期间可以修改配置中的内容，以观察自动定时更新是否生效
		time.Sleep(time.Second * 3)
	}
}
