// 如果你使用示例代码运行
//
// 请在 Consul 中准备对应的配置 spanic/dev:
//
// pre:
//  parh: there-is-pre
// runtime:
//  path: there-is-runtime
//
// 然后启动 Consul 服务
//
// 最后运行示例代码即可
package config

import (
	"context"
	"fmt"
	"log"
	"time"
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

func ExampleConsulConf() {
	ctx := context.Background()
	bindConf := BindConf{} // 选择符合配置格式的结构体
	// 初始化配置对象
	err := InitConf(ctx, Consul, Option{
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
