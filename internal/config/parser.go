package config

// Parser defines the interface for configuration sources.
type Parser interface {
	Parse(cfg *Config) error
}

// MultiParser coordinates multiple parsers in a specific priority order.
type MultiParser struct {
	Parsers []Parser
}

// Parse applies all parsers in order.
func (mp *MultiParser) Parse(cfg *Config) error {
	for _, p := range mp.Parsers {
		if err := p.Parse(cfg); err != nil {
			return err
		}
	}
	return nil
}
