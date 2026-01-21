package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"token13/merchant-backend-go/internal/queue/rabbit"
)

type MerchantCreated struct {
	MerchantID    string `json:"merchant_id"`
	WalletAddress string `json:"wallet_address"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	CreatedAt     string `json:"created_at"`
}

func main() {
	rabbitURL := os.Getenv("RABBIT_URL")
	if rabbitURL == "" {
		log.Fatal("RABBIT_URL is required")
	}
	exchange := os.Getenv("RABBIT_EXCHANGE")
	if exchange == "" {
		exchange = "token13.events"
	}

	conn, err := rabbit.Connect(rabbitURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		log.Fatal(err)
	}

	// durable queue (stable name)
	q, err := ch.QueueDeclare("merchant.created.q", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	if err := ch.QueueBind(q.Name, "merchant.created", exchange, false, nil); err != nil {
		log.Fatal(err)
	}

	// Prefetch 1 = safer while developing
	if err := ch.Qos(1, 0, false); err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "worker-1", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("✅ worker started. waiting for merchant.created ...")

	for m := range msgs {
		var ev MerchantCreated
		if err := json.Unmarshal(m.Body, &ev); err != nil {
			log.Println("❌ bad message:", err)
			_ = m.Nack(false, false) // drop (later send to DLQ)
			continue
		}

		log.Printf("📩 merchant.created: id=%s wallet=%s name=%s email=%s created_at=%s",
			ev.MerchantID, ev.WalletAddress, ev.Name, ev.Email, ev.CreatedAt)

		// TODO: blockchain onboardMerchant(bytes32,address) goes here later.
		time.Sleep(50 * time.Millisecond)

		_ = m.Ack(false)
	}
}
