package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mdp/qrterminal"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/whatsapp"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Server struct {
	Manager *whatsapp.WhatsAppManager
}

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

func RequireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return false
}

func DecodeJSONBody(r *http.Request, target any) error {
	return json.NewDecoder(r.Body).Decode(target)
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) OnboardHandler(w http.ResponseWriter, r *http.Request) {
	if !RequireMethod(w, r, http.MethodPost) {
		return
	}
	var req OnboardRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	device := s.Manager.Container.NewDevice()
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
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, log.Writer())

				png, err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
				if err == nil {
					b64 := base64.StdEncoding.EncodeToString(png)
					log.Printf("Base64 QR (data:image/png;base64): %s", b64)
				}
			} else if evt.Event == "success" {
				log.Printf("Subsystem: Onboarding success for %s. Disconnecting QR client.", req.Email)
				client.Disconnect()

				// Small delay to let WhatsApp state settle.
				time.Sleep(2 * time.Second)

				s.Manager.SpawnInstance(device)
			}
		}
	}()
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Onboarding initiated.")
}

func (s *Server) SendHandler(w http.ResponseWriter, r *http.Request) {
	if !RequireMethod(w, r, http.MethodPost) {
		return
	}
	var req SendRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = domain.Text
	}

	err := s.Manager.SendMessageRequest(req.HostNumber, domain.MessageRequest{
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

func (s *Server) ListBotsHandler(w http.ResponseWriter, r *http.Request) {
	if !RequireMethod(w, r, http.MethodGet) {
		return
	}
	bots := s.Manager.ListInstances()
	WriteJSON(w, http.StatusOK, bots)
}

func (s *Server) GetBotHandler(w http.ResponseWriter, r *http.Request) {
	if !RequireMethod(w, r, http.MethodGet) {
		return
	}
	host := r.URL.Query().Get("host")
	if host == "" {
		http.Error(w, "Host number required", http.StatusBadRequest)
		return
	}
	bot, err := s.Manager.GetInstance(host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	WriteJSON(w, http.StatusOK, bot)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
