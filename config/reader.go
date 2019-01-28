package config

import (
	"errors"
	"os"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// ConfigTagName - represents the tag name for config
const ConfigTagName string = "cfg"

// ConfigDefaultTagName - represents the tag name for default config value
const ConfigDefaultTagName string = "cfg_default"

// ReadConfigFromEnv - reads the config from environment variables
// based on the tagnames
func ReadConfigFromEnv(cfg interface{}) error {
	envMap, err := readEnv(cfg)
	if err != nil {
		return err
	}
	return unmarshal(cfg, envMap)
}

func getConfigType(cfg interface{}) (reflect.Type, error) {
	cfgVal := reflect.ValueOf(cfg)
	// check if not pointer
	if cfgVal.Kind() != reflect.Ptr || cfgVal.IsNil() {
		return nil, errors.New("Invalid argument")
	}

	// get the underlying element≤
	for cfgVal.Kind() == reflect.Ptr {
		cfgVal = cfgVal.Elem()
	}

	if cfgVal.Kind() != reflect.Struct {
		return nil, errors.New("Referenced type is not struct")
	}

	return cfgVal.Type(), nil
}

func getFieldTagDetails(field reflect.StructField, tagName string, defaultValue string) string {
	tagValue := field.Tag.Get(tagName)
	if tagValue == "" {
		tagValue = defaultValue
	}
	return tagValue
}

func readFieldFromEnv(field reflect.StructField, envMap map[string]string) {
	envVarName := getFieldTagDetails(field, ConfigTagName, field.Name)
	if envVarName != "" {
		envVarValue := os.Getenv(envVarName)
		if envVarValue == "" {
			envVarValue = getFieldTagDetails(field, ConfigDefaultTagName, "")
		}
		if envVarValue != "" {
			envMap[envVarName] = envVarValue
		}
	}
}

// readEnv - creates a map of the environment variable and its value based on config tags
func readEnv(cfg interface{}) (interface{}, error) {
	cfgType, err := getConfigType(cfg)
	if err != nil {
		return nil, err
	}

	envMap := make(map[string]string)
	for i := 0; i < cfgType.NumField(); i++ {
		field := cfgType.Field(i)
		readFieldFromEnv(field, envMap)
	}
	return envMap, nil
}

func createDecoderConfig(cfg interface{}) *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           cfg,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
		TagName: ConfigTagName}
}

// Unmarshal decodes the env map to the config Struct.
func unmarshal(cfg interface{}, envMap interface{}) error {
	decoderConfig := createDecoderConfig(cfg)
	decoder, _ := mapstructure.NewDecoder(decoderConfig)
	return decoder.Decode(envMap)
}
