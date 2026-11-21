package main

import (
	"fmt"
	"log"
	"time"

	"github.com/codesbysagar/OrderPulse/internal/domain"
	"github.com/codesbysagar/OrderPulse/internal/storage"
)

func main() {
	repo, err := storage.NewTicketRepository("orderpulse.db")
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	ticket, err := domain.NewTicket(0, "Order #201: Crispy Chicken Combo", "alice", 2*time.Minute)
	if err != nil {
		log.Fatal(err)
	}

	if err := repo.Create(ticket); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Saved new ticket to SQLite with ID #%d\n", ticket.ID)

	_ = ticket.Start()
	if err := repo.Update(ticket); err != nil {
		log.Fatal(err)
	}

	activeTickets, err := repo.ListActive()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n=== Active Tickets Loaded from orderpulse.db (%d total) ===\n", len(activeTickets))
	for _, t := range activeTickets {
		fmt.Printf("[#%d] %-32s | Status: %-11s | SLA Left: %v\n", t.ID, t.Title, t.Status, t.RemainingTime().Round(time.Second))
	}
}
