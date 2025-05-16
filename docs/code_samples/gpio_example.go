package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// PinMode definiert die möglichen Modi eines GPIO-Pins
type PinMode int

const (
	Input PinMode = iota
	Output
	InputPullUp
	InputPullDown
)

// String gibt eine lesbare Repräsentation des PinMode zurück
func (m PinMode) String() string {
	switch m {
	case Input:
		return "input"
	case Output:
		return "output"
	case InputPullUp:
		return "input_pullup"
	case InputPullDown:
		return "input_pulldown"
	default:
		return "unknown"
	}
}

// Edge definiert Trigger-Events für GPIO-Pins
type Edge int

const (
	None Edge = iota
	Rising
	Falling
	Both
)

// String gibt eine lesbare Repräsentation der Edge zurück
func (e Edge) String() string {
	switch e {
	case None:
		return "none"
	case Rising:
		return "rising"
	case Falling:
		return "falling"
	case Both:
		return "both"
	default:
		return "unknown"
	}
}

// Pin repräsentiert einen einzelnen GPIO-Pin
type Pin interface {
	// Grundlegende Operationen
	Read() (bool, error)        // Liest den aktuellen Zustand des Pins
	Write(value bool) error     // Setzt den Pin auf high (true) oder low (false)
	SetMode(mode PinMode) error // Ändert den Modus des Pins
	SetEdge(edge Edge) error    // Ändert die Trigger-Bedingung für Events
	Close() error               // Gibt Ressourcen frei

	// Eigenschaften
	Number() int   // Gibt die Pin-Nummer zurück
	Name() string  // Gibt den Namen des Pins zurück
	Mode() PinMode // Gibt den aktuellen Modus zurück

	// Event-basierte Funktionen
	RegisterCallback(func(bool)) error // Registriert Callback für Statusänderungen
	EnableInterrupt() error            // Aktiviert Interrupts
	DisableInterrupt() error           // Deaktiviert Interrupts
}

// Manager ist verantwortlich für die Verwaltung mehrerer GPIO-Pins
type Manager interface {
	// Pin-Verwaltung
	OpenPin(pinNumber int, mode PinMode) (Pin, error)
	ClosePin(pinNumber int) error
	CloseAll() error

	// Pin-Lookup
	GetPin(pinNumber int) (Pin, bool)
	GetPinByName(name string) (Pin, bool)

	// Konfiguration
	LoadConfig(configFile string) error
	AddPinAlias(number int, name string) error
}

// PinConfig enthält die Konfiguration für einen einzelnen GPIO-Pin
type PinConfig struct {
	Number       int    `json:"number"`
	Name         string `json:"name"`
	Mode         string `json:"mode"`
	InitialState *bool  `json:"initial_state,omitempty"`
	Edge         string `json:"edge,omitempty"`
	DebounceMsec int    `json:"debounce_ms,omitempty"`
	Description  string `json:"description,omitempty"`
}

// GPIOConfig enthält die Konfiguration für alle GPIO-Pins
type GPIOConfig struct {
	Pins []PinConfig `json:"pins"`
}

// basePin implementiert gemeinsame Funktionalität für alle Pin-Implementierungen
type basePin struct {
	number     int
	name       string
	mode       PinMode
	edge       Edge
	debounceMs int
	state      bool
	callbacks  []func(bool)
	mu         sync.RWMutex
}

// baseManager implementiert gemeinsame Funktionalität für alle Manager-Implementierungen
type baseManager struct {
	pins      map[int]Pin
	pinNames  map[string]int
	mu        sync.RWMutex
	openFn    func(int, PinMode) (Pin, error)
	closeFn   func(Pin) error
	interrupt bool
}

// NewManager erstellt eine neue Instanz eines Manager
func NewManager(openFn func(int, PinMode) (Pin, error), closeFn func(Pin) error) Manager {
	return &baseManager{
		pins:     make(map[int]Pin),
		pinNames: make(map[string]int),
		openFn:   openFn,
		closeFn:  closeFn,
	}
}

// Implementierung von Manager-Methoden

func (m *baseManager) OpenPin(pinNumber int, mode PinMode) (Pin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pin, exists := m.pins[pinNumber]; exists {
		return pin, nil
	}

	pin, err := m.openFn(pinNumber, mode)
	if err != nil {
		return nil, err
	}

	m.pins[pinNumber] = pin
	return pin, nil
}

func (m *baseManager) ClosePin(pinNumber int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pin, exists := m.pins[pinNumber]
	if !exists {
		return fmt.Errorf("pin %d is not open", pinNumber)
	}

	err := m.closeFn(pin)
	if err != nil {
		return err
	}

	delete(m.pins, pinNumber)
	return nil
}

func (m *baseManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for _, pin := range m.pins {
		if err := m.closeFn(pin); err != nil {
			lastErr = err
		}
	}

	m.pins = make(map[int]Pin)
	return lastErr
}

func (m *baseManager) GetPin(pinNumber int) (Pin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pin, exists := m.pins[pinNumber]
	return pin, exists
}

func (m *baseManager) GetPinByName(name string) (Pin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pinNumber, exists := m.pinNames[name]
	if !exists {
		return nil, false
	}

	pin, exists := m.pins[pinNumber]
	return pin, exists
}

func (m *baseManager) AddPinAlias(number int, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, exists := m.pinNames[name]; exists && existing != number {
		return fmt.Errorf("pin name '%s' already assigned to pin %d", name, existing)
	}

	m.pinNames[name] = number
	return nil
}

func (m *baseManager) LoadConfig(configFile string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var config GPIOConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	for _, pinCfg := range config.Pins {
		// Parse mode
		var mode PinMode
		switch pinCfg.Mode {
		case "input":
			mode = Input
		case "output":
			mode = Output
		case "input_pullup":
			mode = InputPullUp
		case "input_pulldown":
			mode = InputPullDown
		default:
			return fmt.Errorf("invalid pin mode: %s for pin %d", pinCfg.Mode, pinCfg.Number)
		}

		// Open and configure pin
		pin, err := m.OpenPin(pinCfg.Number, mode)
		if err != nil {
			return err
		}

		// Set initial state for output pins
		if mode == Output && pinCfg.InitialState != nil {
			if err := pin.Write(*pinCfg.InitialState); err != nil {
				return err
			}
		}

		// Set edge if specified
		if pinCfg.Edge != "" {
			var edge Edge
			switch pinCfg.Edge {
			case "none":
				edge = None
			case "rising":
				edge = Rising
			case "falling":
				edge = Falling
			case "both":
				edge = Both
			default:
				return fmt.Errorf("invalid edge: %s for pin %d", pinCfg.Edge, pinCfg.Number)
			}

			if err := pin.SetEdge(edge); err != nil {
				return err
			}
		}

		// Add name alias
		if pinCfg.Name != "" {
			if err := m.AddPinAlias(pinCfg.Number, pinCfg.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

// Beispiel-Implementierung für Linux sysfs

type sysfsPinImpl struct {
	basePin
	path          string
	valueFile     *os.File
	directionFile *os.File
	ctx           context.Context
	cancel        context.CancelFunc
}

func (p *sysfsPinImpl) Read() (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Implementation here would read from sysfs
	// For the example, we just return the stored state
	return p.state, nil
}

func (p *sysfsPinImpl) Write(value bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Implementation here would write to sysfs
	// For the example, we just store the state
	p.state = value
	return nil
}

func (p *sysfsPinImpl) SetMode(mode PinMode) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Implementation here would write to sysfs direction file
	p.mode = mode
	return nil
}

func (p *sysfsPinImpl) SetEdge(edge Edge) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Implementation here would write to sysfs edge file
	p.edge = edge
	return nil
}

func (p *sysfsPinImpl) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
	}

	// Close files and unexport pin
	return nil
}

func (p *sysfsPinImpl) Number() int {
	return p.number
}

func (p *sysfsPinImpl) Name() string {
	return p.name
}

func (p *sysfsPinImpl) Mode() PinMode {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.mode
}

func (p *sysfsPinImpl) RegisterCallback(cb func(bool)) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.callbacks = append(p.callbacks, cb)
	return nil
}

func (p *sysfsPinImpl) EnableInterrupt() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctx != nil {
		return nil // Already running
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())

	// Start a goroutine to monitor the value file for changes
	go func() {
		// Real implementation would use epoll
		// This is just a simplistic example
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var lastState bool
		for {
			select {
			case <-p.ctx.Done():
				return
			case <-ticker.C:
				state, err := p.Read()
				if err != nil {
					continue
				}

				if state != lastState {
					lastState = state
					// Notify callbacks
					for _, cb := range p.callbacks {
						cb(state)
					}
				}
			}
		}
	}()

	return nil
}

func (p *sysfsPinImpl) DisableInterrupt() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
		p.ctx = nil
		p.cancel = nil
	}

	return nil
}

// Beispiel für die Verwendung der GPIO-Schnittstelle

func ExampleUsage() {
	// Erstelle einen GPIO-Manager (zum Beispiel für sysfs)
	gpioManager := NewManager(
		func(pinNumber int, mode PinMode) (Pin, error) {
			// Hier würde die sysfs-spezifische Implementierung sein
			return &sysfsPinImpl{
				basePin: basePin{
					number: pinNumber,
					mode:   mode,
				},
			}, nil
		},
		func(p Pin) error {
			// Schließen des Pins
			return p.Close()
		},
	)

	// Lade Konfiguration aus JSON-Datei
	err := gpioManager.LoadConfig("/etc/owipex/gpio_config.json")
	if err != nil {
		fmt.Printf("Fehler beim Laden der GPIO-Konfiguration: %v\n", err)
		return
	}

	// Zugriff auf einen Pin über seinen Namen
	pumpRelayPin, found := gpioManager.GetPinByName("pump_relay")
	if !found {
		fmt.Println("Pumpen-Relais-Pin nicht gefunden")
		return
	}

	// Schalte das Relais ein
	err = pumpRelayPin.Write(true)
	if err != nil {
		fmt.Printf("Fehler beim Setzen des Pumpen-Relais: %v\n", err)
		return
	}

	// Lese den Button-Zustand
	powerButtonPin, found := gpioManager.GetPinByName("power_button")
	if !found {
		fmt.Println("Power-Button-Pin nicht gefunden")
		return
	}

	// Registriere einen Callback für Änderungen am Button
	powerButtonPin.RegisterCallback(func(state bool) {
		if !state { // Button gedrückt (low bei pull-up)
			fmt.Println("Power-Button wurde gedrückt!")
			// Hier würde die Logik für den Knopfdruck kommen
		}
	})

	// Aktiviere Interrupts für den Button
	err = powerButtonPin.EnableInterrupt()
	if err != nil {
		fmt.Printf("Fehler beim Aktivieren des Button-Interrupts: %v\n", err)
		return
	}

	// Die Anwendung würde hier weiterlaufen...
	time.Sleep(60 * time.Second)

	// Beim Beenden der Anwendung
	gpioManager.CloseAll()
}
