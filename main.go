package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/mdp/qrterminal"
	"github.com/raufhm/whatsapp-testing/config"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/store"
	"github.com/raufhm/whatsapp-testing/whatsapp"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var wm *whatsapp.WhatsAppManager

type OnboardRequest struct {
	Email string `json:"email"`
}

type SendRequest struct {
	HostNumber     string             `json:"host_number"`
	Recipient      string             `json:"recipient"`
	Message        string             `json:"message"`
	IsGroup        bool               `json:"is_group"`
	Type           domain.MessageType `json:"type"`
	MediaPath      string             `json:"media_path,omitempty"`
	ReactionTarget string             `json:"reaction_target,omitempty"`
}

func onboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req OnboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	device := wm.Container.NewDevice()
	client := whatsmeow.NewClient(device, waLog.Stdout("Onboard", "WARN", true))
	qrChan, _ := client.GetQRChannel(context.Background())
	if err := client.Connect(); err != nil {
		http.Error(w, "Failed to connect", http.StatusInternalServerError)
		return
	}
	go func() {
		for evt := range qrChan {
			if evt.Event == "code" {
				log.Printf("Subsystem: New QR Code for %s", req.Email)
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)

				png, err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
				if err == nil {
					b64 := base64.StdEncoding.EncodeToString(png)
					log.Printf("Base64 QR (data:image/png;base64): %s", b64)
				}
			} else if evt.Event == "success" {
				log.Printf("Subsystem: Onboarding success for %s. Disconnecting QR client.", req.Email)
				client.Disconnect()

				// Small delay to let WhatsApp state settle
				time.Sleep(2 * time.Second)

				wm.SpawnInstance(device)
			}
		}
	}()
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Onboarding initiated.")
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = domain.Text
	}

	err := wm.SendMessageRequest(req.HostNumber, domain.MessageRequest{
		Recipient:      req.Recipient,
		Message:        req.Message,
		IsGroup:        req.IsGroup,
		Type:           req.Type,
		MediaPath:      req.MediaPath,
		ReactionTarget: req.ReactionTarget,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Message queued.")
}

func listBotsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	bots := wm.ListInstances()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bots)
}

func getBotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	host := r.URL.Query().Get("host")
	if host == "" {
		http.Error(w, "Host number required", http.StatusBadRequest)
		return
	}
	bot, err := wm.GetInstance(host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bot)
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	pgStore, err := store.NewPostgresStore(cfg.PG_DSN)
	if err != nil {
		log.Fatalf("Store error: %v", err)
	}

	dispatcher := whatsapp.NewMultiDispatcher(
		&whatsapp.LoggerDispatcher{},
		pgStore,
	)

	container, err := sqlstore.New(context.Background(), "pgx", cfg.PG_DSN, waLog.Stdout("DB", "WARN", true))
	if err != nil {
		log.Fatalf("Container error: %v", err)
	}

	wm = whatsapp.NewWhatsAppManager(container, dispatcher, pgStore)
	if err := wm.Start(); err != nil {
		log.Printf("Manager start error: %v", err)
	}

	http.HandleFunc("/api/onboard", onboardHandler)
	http.HandleFunc("/api/send", sendHandler)
	http.HandleFunc("/api/bots", listBotsHandler)
	http.HandleFunc("/api/bots/detail", getBotHandler)

	log.Printf("Server starting on port %s...", cfg.PORT)
	if err := http.ListenAndServe(":"+cfg.PORT, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
