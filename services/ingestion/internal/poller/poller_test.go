package poller

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ecos/ingestion/internal/model"
)

// mockProvider implements Provider for testing.
type mockProvider struct {
	name   string
	result *ProviderResult
	err    error
	calls  int
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Poll(_ context.Context) (*ProviderResult, error) {
	m.calls++
	return m.result, m.err
}

// mockReadingsStore implements store.ReadingsStore for testing.
type mockReadingsStore struct {
	mu       sync.Mutex
	readings []model.SensorReading
	putErr   error
}

func (m *mockReadingsStore) Put(_ context.Context, r model.SensorReading) error {
	if m.putErr != nil {
		return m.putErr
	}
	m.mu.Lock()
	m.readings = append(m.readings, r)
	m.mu.Unlock()
	return nil
}

func (m *mockReadingsStore) GetByStation(_ context.Context, _ string, _ model.ReadingsQuery) ([]model.SensorReading, error) {
	return nil, nil
}

func (m *mockReadingsStore) GetByReadingType(_ context.Context, _ model.ReadingType, _ model.ReadingsQuery) ([]model.SensorReading, error) {
	return nil, nil
}

func (m *mockReadingsStore) List(_ context.Context, _ model.ReadingsQuery) ([]model.SensorReading, error) {
	return nil, nil
}

// mockStationsStore implements store.StationsStore for testing.
type mockStationsStore struct {
	mu       sync.Mutex
	stations []model.Station
	putErr   error
}

func (m *mockStationsStore) Get(_ context.Context, _ string) (*model.Station, error) {
	return nil, nil
}

func (m *mockStationsStore) Put(_ context.Context, s model.Station) error {
	if m.putErr != nil {
		return m.putErr
	}
	m.mu.Lock()
	m.stations = append(m.stations, s)
	m.mu.Unlock()
	return nil
}

func (m *mockStationsStore) Delete(_ context.Context, _ string) error { return nil }

func (m *mockStationsStore) List(_ context.Context) ([]model.Station, error) {
	return nil, nil
}

// mockPublisher implements events.Publisher for testing.
type mockPublisher struct {
	mu     sync.Mutex
	events []model.DataEvent
	err    error
}

func (m *mockPublisher) PublishDataEvent(_ context.Context, e model.DataEvent) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	m.events = append(m.events, e)
	m.mu.Unlock()
	return nil
}

func (m *mockPublisher) PublishAlertEvent(_ context.Context, _ model.AlertEvent) error {
	return nil
}

func TestPoller_ProcessesResult(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		result: &ProviderResult{
			Stations: []model.Station{
				{ID: "test-1", Name: "Test Station"},
			},
			Readings: []model.SensorReading{
				{
					ID:          "r1",
					StationID:   "test-1",
					ReadingType: model.ReadingTypeTemperature,
					Value:       20.5,
					Unit:        "celsius",
				},
				{
					ID:          "r2",
					StationID:   "test-1",
					ReadingType: model.ReadingTypeHumidity,
					Value:       65.0,
					Unit:        "percent",
				},
			},
		},
	}

	readingsStore := &mockReadingsStore{}
	stationsStore := &mockStationsStore{}
	publisher := &mockPublisher{}

	p := New([]Provider{provider}, readingsStore, stationsStore, publisher, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	// Run a single tick.
	p.tick(ctx)
	cancel()

	if provider.calls != 1 {
		t.Errorf("expected 1 poll call, got %d", provider.calls)
	}
	if len(stationsStore.stations) != 1 {
		t.Errorf("expected 1 station stored, got %d", len(stationsStore.stations))
	}
	if len(readingsStore.readings) != 2 {
		t.Errorf("expected 2 readings stored, got %d", len(readingsStore.readings))
	}
	if len(publisher.events) != 2 {
		t.Errorf("expected 2 events published, got %d", len(publisher.events))
	}

	for _, e := range publisher.events {
		if e.EventType != model.EventNewDataReceived {
			t.Errorf("event type = %q, want %q", e.EventType, model.EventNewDataReceived)
		}
	}
}

func TestPoller_CachesKnownStations(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		result: &ProviderResult{
			Stations: []model.Station{
				{ID: "test-1", Name: "Test Station"},
			},
		},
	}

	stationsStore := &mockStationsStore{}
	readingsStore := &mockReadingsStore{}
	publisher := &mockPublisher{}

	p := New([]Provider{provider}, readingsStore, stationsStore, publisher, time.Hour)

	ctx := context.Background()
	// Two ticks — station should only be stored once.
	p.tick(ctx)
	p.tick(ctx)

	if len(stationsStore.stations) != 1 {
		t.Errorf("expected 1 station stored (cached), got %d", len(stationsStore.stations))
	}
}

func TestPoller_SkipsProviderOnError(t *testing.T) {
	badProvider := &mockProvider{
		name: "bad",
		err:  errors.New("connection failed"),
	}
	goodProvider := &mockProvider{
		name: "good",
		result: &ProviderResult{
			Readings: []model.SensorReading{
				{ID: "r1", StationID: "s1"},
			},
		},
	}

	readingsStore := &mockReadingsStore{}
	stationsStore := &mockStationsStore{}
	publisher := &mockPublisher{}

	p := New([]Provider{badProvider, goodProvider}, readingsStore, stationsStore, publisher, time.Hour)

	ctx := context.Background()
	p.tick(ctx)

	if len(readingsStore.readings) != 1 {
		t.Errorf("expected 1 reading (bad provider skipped), got %d", len(readingsStore.readings))
	}
}

func TestPoller_ContinuesOnReadingStoreError(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		result: &ProviderResult{
			Readings: []model.SensorReading{
				{ID: "r1", StationID: "s1"},
				{ID: "r2", StationID: "s1"},
			},
		},
	}

	callCount := 0
	readingsStore := &mockReadingsStore{}
	// Fail on first put only.
	origPut := readingsStore.Put
	_ = origPut
	failOnce := &failOnceReadingsStore{inner: readingsStore}

	stationsStore := &mockStationsStore{}
	publisher := &mockPublisher{}

	p := New([]Provider{provider}, failOnce, stationsStore, publisher, time.Hour)

	ctx := context.Background()
	p.tick(ctx)
	_ = callCount

	// First reading fails, second succeeds.
	if len(publisher.events) != 1 {
		t.Errorf("expected 1 event (first reading failed), got %d", len(publisher.events))
	}
}

// failOnceReadingsStore fails the first Put, then delegates to inner.
type failOnceReadingsStore struct {
	inner *mockReadingsStore
	mu    sync.Mutex
	count int
}

func (f *failOnceReadingsStore) Put(ctx context.Context, r model.SensorReading) error {
	f.mu.Lock()
	f.count++
	n := f.count
	f.mu.Unlock()
	if n == 1 {
		return errors.New("transient error")
	}
	return f.inner.Put(ctx, r)
}

func (f *failOnceReadingsStore) GetByStation(ctx context.Context, id string, q model.ReadingsQuery) ([]model.SensorReading, error) {
	return f.inner.GetByStation(ctx, id, q)
}

func (f *failOnceReadingsStore) GetByReadingType(ctx context.Context, rt model.ReadingType, q model.ReadingsQuery) ([]model.SensorReading, error) {
	return f.inner.GetByReadingType(ctx, rt, q)
}

func (f *failOnceReadingsStore) List(ctx context.Context, q model.ReadingsQuery) ([]model.SensorReading, error) {
	return f.inner.List(ctx, q)
}

func TestPoller_RunStopsOnCancel(t *testing.T) {
	provider := &mockProvider{
		name:   "test",
		result: &ProviderResult{},
	}

	p := New([]Provider{provider}, &mockReadingsStore{}, &mockStationsStore{}, &mockPublisher{}, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		p.Run(ctx)
		close(done)
	}()

	// Let it run at least one tick.
	time.Sleep(80 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success.
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop after cancel")
	}

	if provider.calls < 1 {
		t.Errorf("expected at least 1 poll call, got %d", provider.calls)
	}
}
