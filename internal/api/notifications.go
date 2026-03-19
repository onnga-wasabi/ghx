package api

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
	"github.com/onnga-wasabi/ghx/internal/model"
)

func (c *Client) ListNotifications(ctx context.Context) ([]model.Notification, error) {
	opts := &github.NotificationListOptions{
		All:         true,
		ListOptions: github.ListOptions{PerPage: 50},
	}
	result, _, err := c.GH.Activity.ListNotifications(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	notifs := make([]model.Notification, 0, len(result))
	for _, n := range result {
		notifs = append(notifs, model.Notification{
			ID:        n.GetID(),
			Title:     n.GetSubject().GetTitle(),
			Type:      n.GetSubject().GetType(),
			Reason:    n.GetReason(),
			Unread:    n.GetUnread(),
			RepoName:  n.GetRepository().GetFullName(),
			URL:       n.GetSubject().GetURL(),
			HTMLURL:   n.GetURL(),
			UpdatedAt: n.GetUpdatedAt().Time,
		})
	}
	return notifs, nil
}

func (c *Client) MarkNotificationRead(ctx context.Context, threadID string) error {
	_, err := c.GH.Activity.MarkThreadRead(ctx, threadID)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	return nil
}

func (c *Client) MarkNotificationDone(ctx context.Context, threadID string) error {
	_, err := c.GH.Activity.DeleteThreadSubscription(ctx, threadID)
	if err != nil {
		return fmt.Errorf("mark notification done: %w", err)
	}
	return nil
}
