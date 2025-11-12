package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// UptimeKumaNotification represents the notification from Uptime Kuma
type UptimeKumaNotification struct {
	Heartbeat struct {
		MonitorID int     `json:"monitorID"`
		Status    int     `json:"status"`
		Time      string  `json:"time"`
		Msg       string  `json:"msg"`
		Ping      float64 `json:"ping"`
	} `json:"heartbeat"`
	Monitor struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		URL      string `json:"url"`
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
		Type     string `json:"type"`
	} `json:"monitor"`
	Msg string `json:"msg"`
}

// GoogleChatMessage represents the Card format for Google Chat
type GoogleChatMessage struct {
	Text    string   `json:"text"`
	CardsV2 []CardV2 `json:"cardsV2"`
}

type CardV2 struct {
	CardID string `json:"cardId"`
	Card   Card   `json:"card"`
}

type Card struct {
	Header   CardHeader    `json:"header"`
	Sections []CardSection `json:"sections"`
}

type CardHeader struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	ImageURL string `json:"imageUrl,omitempty"`
}

type CardSection struct {
	Widgets []Widget `json:"widgets"`
}

type Widget struct {
	DecoratedText *DecoratedText `json:"decoratedText,omitempty"`
	TextParagraph *TextParagraph `json:"textParagraph,omitempty"`
	ButtonList    *ButtonList    `json:"buttonList,omitempty"`
}

type DecoratedText struct {
	TopLabel string `json:"topLabel,omitempty"`
	Text     string `json:"text"`
	Icon     *Icon  `json:"icon,omitempty"`
}

type Icon struct {
	KnownIcon string `json:"knownIcon,omitempty"`
}

type TextParagraph struct {
	Text string `json:"text"`
}

type ButtonList struct {
	Buttons []Button `json:"buttons"`
}

type Button struct {
	Text    string   `json:"text"`
	OnClick *OnClick `json:"onClick,omitempty"`
}

type OnClick struct {
	OpenLink *OpenLink `json:"openLink,omitempty"`
}

type OpenLink struct {
	URL string `json:"url"`
}

var googleChatWebhookURL string

func main() {
	// Get Google Chat Webhook URL from environment
	googleChatWebhookURL = os.Getenv("GOOGLE_CHAT_WEBHOOK_URL")
	if googleChatWebhookURL == "" {
		log.Fatal("GOOGLE_CHAT_WEBHOOK_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/health", healthHandler)

	log.Printf("Server starting on port %s", port)
	log.Printf("Forwarding to Google Chat webhook: %s", maskWebhookURL(googleChatWebhookURL))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("Received webhook: %s", string(body))

	// Try to parse as Uptime Kuma notification
	var notification UptimeKumaNotification
	if err := json.Unmarshal(body, &notification); err != nil {
		log.Printf("Error parsing Uptime Kuma notification: %v", err)
		// Send raw message if parsing fails
		sendSimpleMessage(string(body))
	} else {
		// Convert to Google Chat Card format
		chatMessage := convertToGoogleChatCard(notification)
		if err := sendToGoogleChat(chatMessage); err != nil {
			log.Printf("Error sending to Google Chat: %v", err)
			http.Error(w, "Error forwarding message", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func convertToGoogleChatCard(notification UptimeKumaNotification) GoogleChatMessage {
	cleanMsg := sanitizeText(notification.Msg)
	cleanHeartbeatMsg := sanitizeText(notification.Heartbeat.Msg)
	cleanURL := sanitizeText(notification.Monitor.URL)
	cleanHostname := sanitizeText(notification.Monitor.Hostname)

	// Determine status
	isUp := notification.Heartbeat.Status == 1
	statusEmoji := "🔴"
	statusLabel := "Down"
	if isUp {
		statusEmoji = "✅"
		statusLabel = "Up"
	}

	// Build title in Uptime Kuma v1 format: "UP - (monitor name)" or "DOWN - (monitor name)"
	title := fmt.Sprintf("%s - %s", statusLabel, notification.Monitor.Name)

	// Build subtitle from message
	subtitle := cleanMsg
	if subtitle == "" {
		subtitle = cleanHeartbeatMsg
	}
	if subtitle == "" {
		if isUp {
			subtitle = "Service is operational"
		} else {
			subtitle = "Service is experiencing issues"
		}
	}

	// Create widgets
	widgets := []Widget{}

	// Message detail if available
	if cleanHeartbeatMsg != "" {
		widgets = append(widgets, Widget{
			TextParagraph: &TextParagraph{
				Text: cleanHeartbeatMsg,
			},
		})
	}

	// URL
	displayURL := cleanURL
	if displayURL == "" && cleanHostname != "" {
		displayURL = cleanHostname
	}
	if displayURL != "" {
		widgets = append(widgets, Widget{
			DecoratedText: &DecoratedText{
				TopLabel: "URL",
				Text:     displayURL,
			},
		})
	}

	// Response time
	if notification.Heartbeat.Ping > 0 {
		widgets = append(widgets, Widget{
			DecoratedText: &DecoratedText{
				TopLabel: "Response Time",
				Text:     fmt.Sprintf("%.2f ms", notification.Heartbeat.Ping),
			},
		})
	}

	// Time
	if notification.Heartbeat.Time != "" {
		widgets = append(widgets, Widget{
			DecoratedText: &DecoratedText{
				TopLabel: "Time",
				Text:     notification.Heartbeat.Time,
			},
		})
	}

	// Add button to visit URL if available
	if cleanURL != "" {
		widgets = append(widgets, Widget{
			ButtonList: &ButtonList{
				Buttons: []Button{
					{
						Text: "Visit Site",
						OnClick: &OnClick{
							OpenLink: &OpenLink{
								URL: cleanURL,
							},
						},
					},
				},
			},
		})
	}

	// Create detailed preview text for mobile notifications
	var previewLines []string

	// Check if message contains certificate expiration info
	messageToCheck := cleanMsg
	if messageToCheck == "" {
		messageToCheck = cleanHeartbeatMsg
	}
	isCertificateMessage := strings.Contains(strings.ToLower(messageToCheck), "certificate") && 
		(strings.Contains(strings.ToLower(messageToCheck), "expire") || 
		 strings.Contains(strings.ToLower(messageToCheck), "expiration"))

	// First line with status or actual message for certificate expiration
	if isCertificateMessage && messageToCheck != "" {
		// Remove "Down -" or "DOWN -" prefix from certificate message
		cleanedCertMsg := strings.TrimSpace(messageToCheck)
		cleanedCertMsg = strings.TrimPrefix(cleanedCertMsg, "Down -")
		cleanedCertMsg = strings.TrimPrefix(cleanedCertMsg, "DOWN -")
		cleanedCertMsg = strings.TrimPrefix(cleanedCertMsg, "down -")
		cleanedCertMsg = strings.TrimSpace(cleanedCertMsg)
		// Use the cleaned certificate message as the first line (and only line)
		previewLines = append(previewLines, cleanedCertMsg)
	} else {
		// Use generic status message
		if isUp {
			previewLines = append(previewLines, fmt.Sprintf("%s Application is back online", statusEmoji))
		} else {
			previewLines = append(previewLines, fmt.Sprintf("%s Application went down", statusEmoji))
		}

		// Monitor name
		previewLines = append(previewLines, notification.Monitor.Name)

		// Status line with monitor name and emoji
		previewLines = append(previewLines, fmt.Sprintf("[%s] [%s %s]", notification.Monitor.Name, statusEmoji, statusLabel))

		// Add detailed message if available
		if cleanHeartbeatMsg != "" {
			previewLines = append(previewLines, cleanHeartbeatMsg)
		}
	}

	previewText := strings.Join(previewLines, "\n")

	return GoogleChatMessage{
		Text: previewText,
		CardsV2: []CardV2{
			{
				CardID: fmt.Sprintf("uptime-kuma-%d-%d", notification.Monitor.ID, time.Now().Unix()),
				Card: Card{
					Header: CardHeader{
						Title:    title,
						Subtitle: subtitle,
					},
					Sections: []CardSection{
						{
							Widgets: widgets,
						},
					},
				},
			},
		},
	}
}

func sanitizeText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	upper := strings.ToUpper(trimmed)
	if upper == "N/A" || upper == "NA" || upper == "NULL" {
		return ""
	}
	return trimmed
}

func sendToGoogleChat(message GoogleChatMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}

	log.Printf("Sending to Google Chat: %s", string(jsonData))

	resp, err := http.Post(googleChatWebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Println("Successfully sent to Google Chat")
	return nil
}

func sendSimpleMessage(text string) error {
	simpleMsg := map[string]string{"text": text}
	jsonData, err := json.Marshal(simpleMsg)
	if err != nil {
		return err
	}

	resp, err := http.Post(googleChatWebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func maskWebhookURL(url string) string {
	if len(url) < 20 {
		return "***"
	}
	return url[:20] + "***"
}
