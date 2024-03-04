package types

type Dependabot struct {
	Version int `yaml:"version"`
	Updates []struct {
		PackageEcosystem string `yaml:"package-ecosystem"`
		Directory        string `yaml:"directory"`
		Schedule         struct {
			Interval string `yaml:"interval"`
		} `yaml:"schedule"`
	} `yaml:"updates"`
}
