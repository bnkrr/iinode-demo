package main

// @Title        ioconnectorfile.go
// @Description  实现基于文件的输入输出方式（便于调试）
// @Create       dlchang (2024/04/03 15:00)

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type InputMessageFile struct {
	text string
}

func (m *InputMessageFile) Msg() string {
	return m.text
}

func (m *InputMessageFile) Ack() {
}

type IOConnectorFile struct {
	inputOnce  sync.Once
	outputOnce sync.Once
	wgroup     *sync.WaitGroup
	config     *ConfigRunner
	inputCh    chan InputMessage
	outputCh   chan string
}

func (i *IOConnectorFile) InputInit(ctx context.Context) {
	file, err := os.Open(i.config.FileInputPath)

	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	var lines []string
	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	i.inputCh = make(chan InputMessage)
	i.wgroup.Add(1)

	go func() {
		defer i.wgroup.Done()

		for cnt := 0; i.config.FileRestartCount <= 0 || cnt < i.config.FileRestartCount; cnt++ {
			for _, line := range lines {
				select {
				case <-ctx.Done():
					file.Close()
					return
				default:
					i.inputCh <- &InputMessageFile{text: line}
				}
			}
			if i.config.FileRestartInterval > 0 {
				time.Sleep(time.Duration(i.config.FileRestartInterval) * time.Second)
			}
		}
	}()
}

func (i *IOConnectorFile) InputChannel(ctx context.Context) chan InputMessage {
	i.inputOnce.Do(func() { i.InputInit(ctx) })
	return i.inputCh
}

func (i *IOConnectorFile) OutputInit(ctx context.Context) {
	i.outputCh = make(chan string)
	i.wgroup.Add(1)

	if i.config.FileOutputPath == "" {
		go func() {
			defer i.wgroup.Done()

			for {
				select {
				case <-ctx.Done():
					// close output connection
					return
				case output := <-i.outputCh:
					log.Printf("output: %s\n", output)

				}
			}
		}()
	} else {
		file, err := os.Create(i.config.FileOutputPath)
		if err != nil {
			log.Panicf("err create file, %v", err)
		}

		writer := bufio.NewWriter(file)

		go func() {
			defer i.wgroup.Done()

			for {
				select {
				case <-ctx.Done():
					file.Close()
					return
				case output := <-i.outputCh:
					log.Printf("output: %s\n", output)
					_, err := writer.WriteString(output + "\n")
					if err != nil {
						log.Printf("write error %v", err)
					}
					err = writer.Flush()
					if err != nil {
						log.Printf("flush error %v", err)
					}
				}
			}
		}()
	}

}

func (i *IOConnectorFile) OutputChannel(ctx context.Context) chan string {
	i.outputOnce.Do(func() { i.OutputInit(ctx) })
	return i.outputCh
}

func NewIOConnectorFile(config *ConfigRunner, wgroup *sync.WaitGroup) (*IOConnectorFile, error) {
	return &IOConnectorFile{
		inputOnce:  sync.Once{},
		outputOnce: sync.Once{},
		config:     config,
		wgroup:     wgroup,
	}, nil
}
