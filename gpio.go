package gpio

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/warthog618/gpiod"
)

func BoolToInt(v bool) (ret int) {
	if v {
		ret = 1
	}
	return
}

func IntToBool(v int) (ret bool) {
	if v > 0 {
		ret = true
	}
	return
}

type Output struct {
	line *gpiod.Line
}

type SetOutput struct {
	name     string
	value    bool
	response chan error
}

type GPIO struct {
	mu            sync.Mutex
	output        map[string]Output
	CH_SET_OUTPUT chan SetOutput
}

type ConfigGPIO struct {
}

func NewGPIO(output map[string]Output) *GPIO {
	gpio := &GPIO{
		output:        output,
		CH_SET_OUTPUT: make(chan SetOutput, 1),
	}

	return gpio
}

func (g *GPIO) SetOutput(set SetOutput) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if gpio, ok := g.output[set.name]; ok {
		if gpio.line == nil {
			return fmt.Errorf("output name:%s error config", set.name)
		}
		return gpio.line.SetValue(BoolToInt(set.value))
	} else {
		return fmt.Errorf("not found config output name:%s", set.name)
	}
	//return nil
}

func (g *GPIO) Run(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	defer fmt.Println("GPIO.Run Done")
	for {
		select {
		case <-ctx.Done():
			return
		case set := <-g.CH_SET_OUTPUT:
			fmt.Println("set output", set)
			err := g.SetOutput(set)
			if err != nil {
				log.Printf("error set_output %v\n", set)
			}
			func() {
				if err := recover(); err != nil {
					log.Printf("error GPIO write close channel %s", err)
				}
				set.response <- err
			}()
		}
	}
}
