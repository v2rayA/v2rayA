package main

import (
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
)

func TestSubscriptionShouldUpdate(t *testing.T) {
	now := time.Date(2026, time.August, 11, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name         string
		subscription configure.SubscriptionRaw
		serviceStart bool
		want         bool
	}{
		{
			name:         "disabled",
			subscription: configure.SubscriptionRaw{AutoUpdateMode: configure.NotAutoUpdate},
			serviceStart: true,
		},
		{
			name:         "start mode on service start",
			subscription: configure.SubscriptionRaw{AutoUpdateMode: configure.AutoUpdate},
			serviceStart: true,
			want:         true,
		},
		{
			name:         "start mode on periodic check",
			subscription: configure.SubscriptionRaw{AutoUpdateMode: configure.AutoUpdate},
		},
		{
			name: "interval without previous attempt",
			subscription: configure.SubscriptionRaw{
				AutoUpdateMode:         configure.AutoUpdateAtIntervals,
				AutoUpdateIntervalHour: 12,
			},
			want: true,
		},
		{
			name: "invalid interval",
			subscription: configure.SubscriptionRaw{
				AutoUpdateMode: configure.AutoUpdateAtIntervals,
			},
		},
		{
			name: "interval not due",
			subscription: configure.SubscriptionRaw{
				AutoUpdateMode:         configure.AutoUpdateAtIntervals,
				AutoUpdateIntervalHour: 12,
				LastUpdateAttemptAt:    now.Add(-11 * time.Hour).Format(time.RFC3339),
			},
		},
		{
			name: "interval due",
			subscription: configure.SubscriptionRaw{
				AutoUpdateMode:         configure.AutoUpdateAtIntervals,
				AutoUpdateIntervalHour: 12,
				LastUpdateAttemptAt:    now.Add(-12 * time.Hour).Format(time.RFC3339),
			},
			want: true,
		},
		{
			name: "invalid previous attempt is due",
			subscription: configure.SubscriptionRaw{
				AutoUpdateMode:         configure.AutoUpdateAtIntervals,
				AutoUpdateIntervalHour: 12,
				LastUpdateAttemptAt:    "invalid",
			},
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := subscriptionShouldUpdate(test.subscription, now, test.serviceStart); got != test.want {
				t.Fatalf("subscriptionShouldUpdate() = %v, want %v", got, test.want)
			}
		})
	}
}
