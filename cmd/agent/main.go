package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/xKARASb/Calculator/internal/agent"
	"github.com/xKARASb/Calculator/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	computingPower := cfg.ComputingPower

	var wg sync.WaitGroup
	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			a := agent.NewAgent(id, cfg.OrchestratorHost)
			fmt.Println("Started Agent:", id)
			err := a.CalculateExpression()
			if err != nil {
				log.Println("Agent", a.ID, ":", err)
			}
		}(i)
		time.Sleep(500 * time.Millisecond)
	}
	wg.Wait()
}
