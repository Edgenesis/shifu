// package main

// import (
// 	"time"

// 	"github.com/edgenesis/shifu/pkg/gateway/gatewaylwm2m/lwm2m"
// )

// func main() {
// 	client, err := lwm2m.NewClient("10.20.30.173:5683", "test1")
// 	if err != nil {
// 		panic(err)
// 	}

// 	obj := lwm2m.NewObject("3303", nil)
// 	obj.AddObject("0/5700", &lwm2m.Demo1{Data: 123})
// 	obj.AddObject("0/5601", &lwm2m.Demo1{Data: 123})
// 	obj.AddObject("0/5701", &lwm2m.Demo{Str: "false"})
// 	obj.AddObject("0/5750", &lwm2m.Demo{Str: "123"})
// 	client.AddObject(*obj)

// 	client.Register()

// 	time.Sleep(time.Second * 20)
// 	obj.Id = "3304"
// 	client.AddObject(*obj)

// 	select {}
// }

package main

import (
	"github.com/edgenesis/shifu/pkg/gateway/gatewaylwm2m"
	"github.com/edgenesis/shifu/pkg/logger"
)

func main() {
	client, err := gatewaylwm2m.New()
	if err != nil {
		logger.Fatal(err)
	}

	err = client.LoadCfg()
	if err != nil {
		logger.Fatal(err)
	}

	panic(client.Start())
}
