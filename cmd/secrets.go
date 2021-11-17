package cmd

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func ReadSecretsFromYaml(yamlFile string) *Secrets {
	var secrets map[string]string
	file, err := os.Open(yamlFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to open secrets file %v. Cause: %v.", yamlFile, err.Error()))
	}
	reader := bufio.NewReader(file)
	err = yaml.NewDecoder(reader).Decode(&secrets)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse secrets file %v. Cause %v.", yamlFile, err.Error()))
	}
	return &Secrets{secrets: secrets}
}

type Secrets struct {
	secrets map[string]string
}

func (resolver *Secrets) resolveSecret(secretName string) string {
	secret, found := resolver.secrets[secretName]
	if !found {
		panic(fmt.Sprintf("Missing value for secret %v", secretName))
	}
	return secret
}
