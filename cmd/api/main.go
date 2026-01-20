package main

import (
	"log"

	"token13/merchant-backend-go/internal/app"
)

func main() {
	c, err := app.Wire()
	if err != nil {
		log.Fatal(err)
	}

	addr := app.Addr(c.Cfg.HTTPPort)
	c.Log.Info("api_starting", "addr", addr, "env", c.Cfg.AppEnv)

	if err := c.API.Engine.Run(addr); err != nil {
		c.Log.Error("api_failed", "err", err)
		log.Fatal(err)
	}
}
