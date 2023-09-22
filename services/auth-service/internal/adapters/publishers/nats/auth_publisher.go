package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"bangungrid/services/auth-service/internal/core/domain"
	"bangungrid/services/auth-service/internal/core/ports"
)

type AuthNatsPublisher struct {
	js nats.JetStreamContext
}

func NewAuthNatsPublisher(js nats.JetStreamContext) ports.AuthEventPublisherPort {
	if js != nil {
		// Pastikan stream AUTH dengan subject 'bangungrid.auth.>' sudah terdaftar
		_, err := js.StreamInfo("AUTH")
		if err != nil {
			_, err = js.AddStream(&nats.StreamConfig{
				Name:      "AUTH",
				Subjects:  []string{"bangungrid.auth.>"},
				Storage:   nats.FileStorage,
				Retention: nats.LimitsPolicy,
			})
			if err != nil {
				log.Printf("[NATS] Peringatan: Gagal membuat stream AUTH: %v\n", err)
			} else {
				log.Println("[NATS] Stream AUTH berhasil diinisialisasi otomatis")
			}
		}
	}
	return &AuthNatsPublisher{js: js}
}

type UserRegisteredEvent struct {
	UserID         string          `json:"user_id"`
	TenantID       string          `json:"tenant_id"`
	OrganizationID string          `json:"organization_id,omitempty"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone"`
	UserType       domain.UserType `json:"user_type"`
	Role           string          `json:"role"`
	RegisteredAt   time.Time       `json:"registered_at"`
}

func (p *AuthNatsPublisher) PublishUserRegistered(ctx context.Context, u *domain.User) error {
	if p.js == nil {
		log.Println("[NATS] JetStream context tidak aktif, lewati publikasi event")
		return nil
	}

	var orgID string
	if u.OrganizationID != nil {
		orgID = *u.OrganizationID
	}

	event := UserRegisteredEvent{
		UserID:         u.ID,
		TenantID:       u.TenantID,
		OrganizationID: orgID,
		Email:          u.Email,
		Phone:          u.Phone,
		UserType:       u.UserType,
		Role:           u.Role,
		RegisteredAt:   u.CreatedAt,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("gagal serialisasi user registered event: %w", err)
	}

	msg := &nats.Msg{
		Subject: "bangungrid.auth.user.registered",
		Data:    data,
		Header: nats.Header{
			"Nats-Msg-Id": []string{fmt.Sprintf("user-reg-%s", u.ID)},
			"Timestamp":   []string{time.Now().UTC().Format(time.RFC3339)},
		},
	}

	_, err = p.js.PublishMsg(msg, nats.Context(ctx))
	if err != nil {
		log.Printf("[NATS] Gagal mempublikasikan event user registered: %v\n", err)
		return err
	}

	log.Printf("[NATS] Berhasil mempublikasikan user registered event: %s\n", u.ID)
	return nil
}
