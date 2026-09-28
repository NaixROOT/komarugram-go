package mockstore

import (
	"context"
	"time"

	"komarugram/internal/messenger/model"
)

// Sessions implements model.SessionsSource with made-up devices.
func (s *Store) Sessions(ctx context.Context) ([]model.Session, error) {
	now := time.Now()
	return []model.Session{
		{Current: true, Official: false, APIID: 12345, Device: "Демо-компьютер", Platform: "Linux", System: "Ubuntu 24.04", App: "komarugram-go 0.1", IP: "192.0.2.10", Country: "Россия", Active: now},
		{Hash: 1, Official: true, APIID: 6, Device: "Pixel 8", Platform: "Android", System: "SDK 35", App: model.SessionApp(6, "Telegram Android", "12.0.1"), IP: "198.51.100.7", Country: "Россия", Active: now.Add(-2 * time.Hour), Created: now.AddDate(0, -3, 0)},
		{Hash: 2, Official: true, APIID: 2040, Device: "Рабочий ноутбук", Platform: "Windows", System: "Windows 11", App: model.SessionApp(2040, "Telegram Desktop", "6001000"), IP: "203.0.113.4", Country: "Казахстан", Active: now.AddDate(0, 0, -2), Created: now.AddDate(-1, 0, 0)},
		{Hash: 3, Official: true, APIID: 2496, Device: "Chrome 140", Platform: "Web", System: "macOS", App: "Telegram Web A 10.9", IP: "203.0.113.9", Country: "Германия", Active: now.AddDate(0, -1, 0), Created: now.AddDate(0, -2, 0)},
		{Hash: 4, Incomplete: true, APIID: 10840, Device: "iPhone 15", Platform: "iOS", System: "iOS 19", App: "Telegram iOS 12.0", IP: "192.0.2.77", Country: "Нидерланды", Active: now.Add(-30 * time.Minute)},
	}, nil
}
