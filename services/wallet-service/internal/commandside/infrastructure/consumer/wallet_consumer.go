package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/commandside/domain"
	"wallet/wallet-service/internal/commandside/infrastructure/persistence"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentDBConsumer struct {
	client    *kurrentdb.Client
	esHandler *persistence.WalletKurrentDBESHandler
}

func NewWalletKurrentDBConsumer(
	client *kurrentdb.Client,
	esHandler *persistence.WalletKurrentDBESHandler,
) *WalletKurrentDBConsumer {
	return &WalletKurrentDBConsumer{
		client:    client,
		esHandler: esHandler,
	}
}

func (consumer *WalletKurrentDBConsumer) Start(ctx context.Context) {
	groupName := config.Load().WalletCommandGroupName
	subscriptionOptions := kurrentdb.SubscribeToPersistentSubscriptionOptions{}
	subscription, err := consumer.client.SubscribeToPersistentSubscriptionToAll(ctx, groupName, subscriptionOptions)
	if err != nil {
		slog.Error("failed to subscribe to persistent subscription", slog.String("group_name", groupName), slog.String("error", err.Error()))
		return
	}
	defer subscription.Close()

	for {
		msg := subscription.Recv()
		if msg.SubscriptionDropped != nil {
			return
		}
		if msg.EventAppeared == nil ||
			msg.EventAppeared.Event == nil ||
			msg.EventAppeared.Event.Event == nil {
			continue
		}
		if err := consumer.handle(
			msg.EventAppeared.Event.Event.Data,
			msg.EventAppeared.Event.Event.EventType,
		); err != nil {
			return
		}
		if err := subscription.Ack(msg.EventAppeared.Event); err != nil {
			return
		}
	}
}

func (consumer *WalletKurrentDBConsumer) handle(
	eventBody []byte,
	eventName string,
) error {
	switch eventName {
	case pkg.FundsTransferredEventName:
		var event pkg.FundsTransferredEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal funds transferred event", slog.String("error", err.Error()))
			return err
		}
		return consumer.receiveFundsTransfer(event)

	default:
		slog.Warn("unhandled event type in wallet consumer", slog.String("type", eventName))
		return nil
	}
}

func (consumer *WalletKurrentDBConsumer) receiveFundsTransfer(event pkg.FundsTransferredEvent) error {
	toWalletID, err := valueobject.NewID(event.ToWalletID)
	if err != nil {
		slog.Error("invalid to_wallet_id in funds transferred event", slog.String("to_wallet_id", event.ToWalletID), slog.String("error", err.Error()))
		return err
	}
	transferID, err := valueobject.NewID(event.TransferID)
	if err != nil {
		slog.Error("invalid transfer_id in funds transferred event", slog.String("transfer_id", event.TransferID), slog.String("error", err.Error()))
		return err
	}
	fromWalletID, err := valueobject.NewID(event.FromWalletID)
	if err != nil {
		slog.Error("invalid from_wallet_id in funds transferred event", slog.String("from_wallet_id", event.FromWalletID), slog.String("error", err.Error()))
		return err
	}
	amount, err := valueobject.NewMoney(event.AmountInCents)
	if err != nil {
		slog.Error("invalid amount in funds transferred event", slog.Int("amount_in_cents", event.AmountInCents), slog.String("error", err.Error()))
		return err
	}

	wallet, found, err := consumer.esHandler.Find(toWalletID)
	if err != nil {
		return err
	}
	if !found {
		slog.Error("wallet not found for funds transfer", slog.String("wallet_id", toWalletID.String()))
		return pkg.ErrWalletNotFound
	}

	command := domain.ReceiveFundsTransferCommand{
		Amount:       amount,
		TransferID:   transferID,
		FromWalletID: fromWalletID,
		Timestamp:    event.Timestamp,
	}
	if err := wallet.ReceiveFundsTransfer(command); err != nil {
		return err
	}
	return consumer.esHandler.Save(wallet)
}
