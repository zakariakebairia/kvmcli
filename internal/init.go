package internal

// import (
// 	"fmt"
// 	"os"
//
// 	"github.com/zakariakebairia/kvmcli/internal"
// 	"github.com/zakariakebairia/kvmcli/internal/logger"
// 	"gopkg.in/yaml.v3"
// )
//
// const configFile = "servers.yml"
//
// func CreateYAMLConfig() error {
// 	config := vms.VirtualMachine{
// 		ApiVersion: "kvmcli/v1",
// 		Kind:       "VirtualMachine",
// 		Metadata: config.Metadata{
// 			Name: "vm-01",
// 			Labels: map[string]string{
// 				"environment": "production",
// 				"role":        "worker",
// 			},
// 		},
// 		Spec: config.Spec{
// 			CPU:    1,
// 			Memory: 1024,
// 			Image:  "rocky9.5",
// 			Disk: config.Disk{
// 				Size: "20G",
// 				Path: "<Path for the image from where the VM will boot>",
// 			},
// 			Network: config.Network{
// 				Name:       "homelab",
// 				MacAddress: "00:00:00:00:00:00",
// 			},
// 			Autostart: true,
// 		},
// 	}
// 	// Marshal the struct to YAML.
// 	yamlData, err := yaml.Marshal(config)
// 	if err != nil {
// 		return fmt.Errorf("error marshalling YAML: %w", err)
// 	}
//
// 	// Write the YAML data to a file.
// 	err = os.WriteFile(configFile, yamlData, 0644)
// 	if err != nil {
// 		return fmt.Errorf("error writing YAML file: %w", err)
// 	}
//
// 	logger.Log.Infof("%s created successfully!", configFile)
// 	logger.Log.Info("Use this file as a template for your cluster")
// 	return nil
// }
