package config

type Project struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Version     string   `toml:"version"`
	Requires    []string `toml:"requires"`
}

type Package struct {
	Include []string `toml:"include"`
	Exclude []string `toml:"exclude"`
}

type ProofmanConfig struct {
	Project Project `toml:"project"`
	Package Package `toml:"package"`
}
