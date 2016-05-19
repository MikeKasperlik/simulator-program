package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	//"sync"
	"time"

	//"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"

	"github.com/carloop/simulator/mcp2515"
	"github.com/carloop/simulator/reportServer"
)

func main() {
	flag.Parse()

	// err := embd.InitSPI()
	// if err != nil {
	// 	panic(err)
	// }
	// defer embd.CloseSPI()

	// const (
	// 	device  = 0
	// 	speed   = 1e5
	// 	bpw     = 8
	// 	delay   = 0
	// 	channel = 0
	// )

	// spi := embd.NewSPIBus(embd.SPIMode0, device, int(speed), bpw, delay)
	// defer spi.Close()

	// canDevice := mcp2515.New(spi)
	// err = canDevice.Setup(500000)

	// if err != nil {
	// 	printError(err)
	// 	return
	// }

	//rxChan := make(mcp2515.MsgChan, 10)
	txChan := make(mcp2515.MsgChan, 10)
	//errChan := make(mcp2515.ErrChan, 10)
	dataChan := make(reportServer.DataChan, 1)

	server := reportServer.New(dataChan)

	// var wg sync.WaitGroup
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	mcp2515.RunMessageLoop(canDevice, rxChan, txChan, errChan)
	// }()
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	printCanMessages(rxChan, txChan, errChan)
	// }()
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	sendMessages(txChan)
	// }()

	go func() {
		forwardDataToCAN(dataChan, txChan)
	}()

	// Don't wait on this gorouting, just close the server when exiting
	// the program
	//go func() {
	server.Start()
	//}()

	// Wait for all goroutines to be done
	// wg.Wait()
}

func printCanMessages(rxChan mcp2515.MsgChan, txChan mcp2515.MsgChan,
	errChan mcp2515.ErrChan) {

	fmt.Println("Starting CAN receiver")

	startTime := time.Now()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		select {
		case rxMessage := <-rxChan:
			printMessage(rxMessage, startTime)
		case err := <-errChan:
			printError(err)
		case <-c:
			// Program done
			return
		}
	}
}

func printMessage(message *mcp2515.Message, startTime time.Time) {
	timeOffset := message.Time.Sub(startTime).Seconds()
	fmt.Printf("%15.6f %03x %d", timeOffset, message.Id, message.Length)
	for i := uint8(0); i < message.Length; i++ {
		fmt.Printf(" %02x", message.Data[i])
	}
	fmt.Println("")

}

func printError(err error) {
	fmt.Printf("Error occured: %v", err)
	fmt.Println("")
}

func sendMessages(txChan mcp2515.MsgChan) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	i := uint8(0)
	for {
		var message mcp2515.Message
		message.Id = 0x2AA
		message.Length = 8
		for j := 0; j < 8; j++ {
			message.Data[j] = 0xAA
		}
		i += 1

		select {
		case txChan <- &message:
			// Message added to queue

		case <-c:
			// Program done
			return
		default:
			// If tx channel is full, ignore
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func forwardDataToCAN(incomingData reportServer.DataChan, txChan mcp2515.MsgChan) {
	for {
		data := <-incomingData
		message := dataToCAN(data)
		txChan <- message
	}
}

func dataToCAN(data reportServer.ReportData) *mcp2515.Message {
	var message mcp2515.Message
	message.Id = 0x55
	message.Length = 8

	speedInt := int(data.Speed * 100)
	latInt := int((data.Lat + 360) * 20000)
	lngInt := int((data.Lng + 360) * 20000)

	message.Data[0] = uint8(speedInt >> 8)
	message.Data[1] = uint8(speedInt & 0xFF)
	message.Data[2] = uint8(latInt >> 16)
	message.Data[3] = uint8((latInt >> 8) & 0xFF)
	message.Data[4] = uint8(latInt & 0xFF)
	message.Data[5] = uint8(lngInt >> 16)
	message.Data[6] = uint8((lngInt >> 8) & 0xFF)
	message.Data[7] = uint8(lngInt & 0xFF)

	return &message
}
