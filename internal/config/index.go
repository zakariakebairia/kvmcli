package config

import "fmt"

type Index struct {
	Networks map[string]struct{}
	Stores   map[string]struct{}
}

func BuildIndex(cfg *Config) (*Index, error) {
	idx := &Index{
		Networks: map[string]struct{}{},
		Stores:   map[string]struct{}{},
	}
	for _, net := range cfg.Networks {
		if net.Name == "" {
			return nil, fmt.Errorf("network with empty name")
		}
		if _, ok := idx.Networks[net.Name]; ok {
			return nil, fmt.Errorf("duplicate network %q", net.Name)
		}
		idx.Networks[net.Name] = struct{}{}
	}

	for _, st := range cfg.Stores {
		if st.Name == "" {
			return nil, fmt.Errorf("store with empty name")
		}
		if _, ok := idx.Stores[st.Name]; ok {
			return nil, fmt.Errorf("duplicate store %q", st.Name)
		}
		idx.Stores[st.Name] = struct{}{}
	}
	return idx, nil
}
