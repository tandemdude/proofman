package config

type Project struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Version     string   `toml:"version"`
	Requires    []string `toml:"requires"`
}

type ProofmanConfig struct {
	Project Project `toml:"project"`
}
