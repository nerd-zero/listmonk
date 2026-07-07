package main

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/listmonk/internal/core"
	"github.com/knadh/listmonk/internal/manager"
)

// tenantMessengers is a manager.MessengerResolver that lazily builds and
// caches SMTP messengers per tenant from that tenant's own settings
// (settings are per-tenant since phase 5). Constructed once and given to
// manager.New(); it holds a reference to Core rather than a snapshot of
// settings, so it always resolves against the DB.
//
// The cache is never invalidated within a process's lifetime - this
// matches the existing behavior of the process-global messenger set
// (built once at boot) and the fact that settings updates already require
// a full process restart to take effect (cmd/settings.go's
// handleSettingsRestart -> syscall.Exec). Live invalidation without a
// restart is out of scope here; see issue #41.
type tenantMessengers struct {
	core *core.Core

	mu    sync.Mutex
	cache map[int]map[string]manager.Messenger
}

func newTenantMessengers(co *core.Core) *tenantMessengers {
	return &tenantMessengers{
		core:  co,
		cache: make(map[int]map[string]manager.Messenger),
	}
}

// GetMessenger implements manager.MessengerResolver.
func (t *tenantMessengers) GetMessenger(ctx context.Context, tenantID int, name string) (manager.Messenger, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if m, ok := t.cache[tenantID]; ok {
		if msgr, ok := m[name]; ok {
			return msgr, nil
		}
		return nil, manager.ErrMessengerNotFound
	}

	settings, err := t.core.GetSettings(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Settings' JSON tags (e.g. "smtp") match the dotted keys initSettings
	// loads from the DB into a koanf instance at boot - round-tripping
	// through JSON here reuses that same shape for a fresh, tenant-scoped
	// koanf instance instead of duplicating the DB-to-koanf loading logic.
	b, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	tenantKo := koanf.New(".")
	if err := tenantKo.Load(confmap.Provider(raw, "."), nil); err != nil {
		return nil, err
	}

	msgrs, err := initSMTPMessengers(tenantKo)
	if err != nil {
		return nil, err
	}

	m := make(map[string]manager.Messenger, len(msgrs))
	for _, msgr := range msgrs {
		m[msgr.Name()] = msgr
	}
	t.cache[tenantID] = m

	if msgr, ok := m[name]; ok {
		return msgr, nil
	}
	return nil, manager.ErrMessengerNotFound
}
