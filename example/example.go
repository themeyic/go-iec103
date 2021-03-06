package main

import (
	"fmt"
	iec "github.com/themeyic/go-iec103"
	"github.com/themeyic/go-iec103/ieccon"
	"time"
)

func main() {
	//调用ClientProvider的构造函数,返回结构体指针
	p := iec.NewClientProvider()
	//windows 下面就是 com开头的，比如说 com3
	//mac OS 下面就是 /dev/下面的，比如说 dev/tty.usbserial-14320
	p.Address = "com2"
	p.BaudRate = 19200
	p.DataBits = 8
	p.Parity = "O"
	p.StopBits = 1
	p.Timeout = 100 * time.Millisecond

	client := ieccon.NewClient(p)
	client.LogMode(true)
	err := client.Start()
	if err != nil {
		fmt.Println("start err,", err)
		return
	}
	test := &iec.Iec103ConfigClient{"01", 0, 1, "15", "2a", "fe", "f1", "09"}
	if test.Initialize(client) == "Loading Finished !" {
		test.SummonSecondaryData(client)
		test1 := []int{1,2,3}
		test2 :=  test.MasterStationReadsAnalogQuantity(client,test1)
		fmt.Println(test2)
	}

}
