/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package viperutil

import (
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"
	version "github.com/hashicorp/go-version"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

// ConfigPaths returns the paths from environment and
// defaults which are CWD and /etc/hyperledger/fabric.
func ConfigPaths() []string {
	var paths []string
	if p := os.Getenv("FABRIC_CFG_PATH"); p != "" {
		paths = append(paths, p)
	}
	return append(paths, ".", "/etc/hyperledger/fabric")
}

// ConfigParser holds the configuration file locations.
// It keeps the config file directory locations and env variables.
// From the file the config is unmarshalled and stored.
// Currently "yaml" is supported.
type ConfigParser struct {
	// configuration file to process
	configPaths []string
	configName  string
	configFile  string

	// parsed config
	config map[string]interface{}
}

// New creates a ConfigParser instance
func New() *ConfigParser {
	return &ConfigParser{
		config: map[string]interface{}{},
	}
}

// AddConfigPaths keeps a list of path to search the relevant
// config file. Multiple paths can be provided.
func (c *ConfigParser) AddConfigPaths(cfgPaths ...string) {
	c.configPaths = append(c.configPaths, cfgPaths...)
}

// SetConfigName provides the configuration file name stem. The upper-cased
// version of this value also serves as the environment variable override
// prefix.
func (c *ConfigParser) SetConfigName(in string) {
	c.configName = in
}

// ConfigFileUsed returns the used configFile.
func (c *ConfigParser) ConfigFileUsed() string {
	return c.configFile
}

// Search for the existence of filename for all supported extensions
func (c *ConfigParser) searchInPath(in string) (filename string) {
	supportedExts := []string{"yaml", "yml"}
	for _, ext := range supportedExts {
		fullPath := filepath.Join(in, c.configName+"."+ext)
		_, err := os.Stat(fullPath)
		if err == nil {
			return fullPath
		}
	}
	return ""
}

// Search for the configName in all configPaths
func (c *ConfigParser) findConfigFile() string {
	paths := c.configPaths
	if len(paths) == 0 {
		paths = ConfigPaths()
	}
	for _, cp := range paths {
		file := c.searchInPath(cp)
		if file != "" {
			return file
		}
	}
	return ""
}

// Get the valid and present config file
func (c *ConfigParser) getConfigFile() string {
	// if explicitly set, then use it
	if c.configFile != "" {
		return c.configFile
	}

	c.configFile = c.findConfigFile()
	return c.configFile
}

// ReadInConfig reads and unmarshals the config file.
func (c *ConfigParser) ReadInConfig() error {
	cf := c.getConfigFile()
	log.Printf("Attempting to open the config file: %s", cf)
	file, err := os.Open(cf)
	if err != nil {
		log.Printf("Unable to open the config file: %s", cf)
		return err
	}
	defer file.Close()

	return c.ReadConfig(file)
}

// ReadConfig parses the buffer and initializes the config.
func (c *ConfigParser) ReadConfig(in io.Reader) error {
	return yaml.NewDecoder(in).Decode(c.config)
}

// Get value for the key by searching environment variables.
func (c *ConfigParser) getFromEnv(key string) string {
	envKey := key
	if c.configName != "" {
		envKey = c.configName + "_" + envKey
	}
	envKey = strings.ToUpper(envKey)
	envKey = strings.ReplaceAll(envKey, ".", "_")
	return os.Getenv(envKey)
}

// Prototype declaration for getFromEnv function.
type envGetter func(key string) string

func getKeysRecursively(base string, getenv envGetter, nodeKeys map[string]interface{}, oType reflect.Type) map[string]interface{} {
	subTypes := map[string]reflect.Type{}

	if oType != nil && oType.Kind() == reflect.Struct {
	outer:
		for i := 0; i < oType.NumField(); i++ {
			fieldName := oType.Field(i).Name
			fieldType := oType.Field(i).Type

			for key := range nodeKeys {
				if strings.EqualFold(fieldName, key) {
					subTypes[key] = fieldType
					continue outer
				}
			}

			subTypes[fieldName] = fieldType
			nodeKeys[fieldName] = nil
		}
	}

	result := make(map[string]interface{})
	for key, val := range nodeKeys {
		fqKey := base + key

		// overwrite val, if an environment is available
		if override := getenv(fqKey); override != "" {
			val = override
		}

		switch val := val.(type) {
		case map[string]interface{}:
			log.Printf("Found map[string]interface{} value for %s", fqKey)
			result[key] = getKeysRecursively(fqKey+".", getenv, val, subTypes[key])

		case map[interface{}]interface{}:
			log.Printf("Found map[interface{}]interface{} value for %s", fqKey)
			result[key] = getKeysRecursively(fqKey+".", getenv, toMapStringInterface(val), subTypes[key])

		case nil:
			if override := getenv(fqKey + ".File"); override != "" {
				result[key] = map[string]interface{}{"File": override}
			}

		default:
			result[key] = val
		}
	}
	return result
}

func toMapStringInterface(m map[interface{}]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range m {
		k, ok := k.(string)
		if !ok {
			panic(fmt.Sprintf("Non string %v, %v: key-entry: %v", k, v, k))
		}
		result[k] = v
	}
	return result
}

// customDecodeHook parses strings of the format "[thing1, thing2, thing3]"
// into string slices. Note that whitespace around slice elements is removed.
func customDecodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}

	raw := data.(string)
	l := len(raw)
	if l > 1 && raw[0] == '[' && raw[l-1] == ']' {
		slice := strings.Split(raw[1:l-1], ",")
		for i, v := range slice {
			slice[i] = strings.TrimSpace(v)
		}
		return slice, nil
	}

	return data, nil
}

func byteSizeDecodeHook(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	if f != reflect.String || t != reflect.Uint32 {
		return data, nil
	}
	raw := data.(string)
	if raw == "" {
		return data, nil
	}
	re := regexp.MustCompile(`^(?P<size>[0-9]+)\s*(?i)(?P<unit>(k|m|g))b?$`)
	if re.MatchString(raw) {
		size, err := strconv.ParseUint(re.ReplaceAllString(raw, "${size}"), 0, 64)
		if err != nil {
			return data, nil
		}
		unit := re.ReplaceAllString(raw, "${unit}")
		switch strings.ToLower(unit) {
		case "g":
			size = size << 10
			fallthrough
		case "m":
			size = size << 10
			fallthrough
		case "k":
			size = size << 10
		}
		if size > math.MaxUint32 {
			return size, fmt.Errorf("value '%s' overflows uint32", raw)
		}
		return size, nil
	}
	return data, nil
}

func stringFromFileDecodeHook(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	// "to" type should be string
	if t != reflect.String {
		return data, nil
	}
	// "from" type should be map
	if f != reflect.Map {
		return data, nil
	}
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.String:
		return data, nil
	case reflect.Map:
		d := data.(map[string]interface{})
		fileName, ok := d["File"]
		if !ok {
			fileName, ok = d["file"]
		}
		switch {
		case ok && fileName != nil:
			bytes, err := os.ReadFile(fileName.(string))
			if err != nil {
				return data, err
			}
			return string(bytes), nil
		case ok:
			// fileName was nil
			return nil, fmt.Errorf("value of File: was nil")
		}
	}
	return data, nil
}

func pemBlocksFromFileDecodeHook(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	// "to" type should be string
	if t != reflect.Slice {
		return data, nil
	}
	// "from" type should be map
	if f != reflect.Map {
		return data, nil
	}
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.String:
		return data, nil
	case reflect.Map:
		var fileName string
		var ok bool
		switch d := data.(type) {
		case map[string]string:
			fileName, ok = d["File"]
			if !ok {
				fileName, ok = d["file"]
			}
		case map[string]interface{}:
			var fileI interface{}
			fileI, ok = d["File"]
			if !ok {
				fileI = d["file"]
			}
			fileName, ok = fileI.(string)
		}

		switch {
		case ok && fileName != "":
			var result []string
			bytes, err := os.ReadFile(fileName)
			if err != nil {
				return data, err
			}
			for len(bytes) > 0 {
				var block *pem.Block
				block, bytes = pem.Decode(bytes)
				if block == nil {
					break
				}
				if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
					continue
				}
				result = append(result, string(pem.EncodeToMemory(block)))
			}
			return result, nil
		case ok:
			// fileName was nil
			return nil, fmt.Errorf("value of File: was nil")
		}
	}
	return data, nil
}

var kafkaVersionConstraints map[sarama.KafkaVersion]version.Constraints

func init() {
	kafkaVersionConstraints = make(map[sarama.KafkaVersion]version.Constraints)
	kafkaVersionConstraints[sarama.V0_8_2_0], _ = version.NewConstraint(">=0.8.2,<0.8.2.1")
	kafkaVersionConstraints[sarama.V0_8_2_1], _ = version.NewConstraint(">=0.8.2.1,<0.8.2.2")
	kafkaVersionConstraints[sarama.V0_8_2_2], _ = version.NewConstraint(">=0.8.2.2,<0.9.0.0")
	kafkaVersionConstraints[sarama.V0_9_0_0], _ = version.NewConstraint(">=0.9.0.0,<0.9.0.1")
	kafkaVersionConstraints[sarama.V0_9_0_1], _ = version.NewConstraint(">=0.9.0.1,<0.10.0.0")
	kafkaVersionConstraints[sarama.V0_10_0_0], _ = version.NewConstraint(">=0.10.0.0,<0.10.0.1")
	kafkaVersionConstraints[sarama.V0_10_0_1], _ = version.NewConstraint(">=0.10.0.1,<0.10.1.0")
	kafkaVersionConstraints[sarama.V0_10_1_0], _ = version.NewConstraint(">=0.10.1.0,<0.10.2.0")
	kafkaVersionConstraints[sarama.V0_10_2_0], _ = version.NewConstraint(">=0.10.2.0,<0.11.0.0")
	kafkaVersionConstraints[sarama.V0_11_0_0], _ = version.NewConstraint(">=0.11.0.0,<1.0.0")
	kafkaVersionConstraints[sarama.V1_0_0_0], _ = version.NewConstraint(">=1.0.0")
}

func kafkaVersionDecodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(sarama.KafkaVersion{}) {
		return data, nil
	}

	v, err := version.NewVersion(data.(string))
	if err != nil {
		return nil, fmt.Errorf("unable to parse Kafka version: %s", err)
	}

	for kafkaVersion, constraints := range kafkaVersionConstraints {
		if constraints.Check(v) {
			return kafkaVersion, nil
		}
	}

	return nil, fmt.Errorf("unsupported Kafka version: '%s'", data)
}

func bccspHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(&factory.FactoryOpts{}) {
		return data, nil
	}

	config := factory.GetDefaultOpts()

	err := mapstructure.WeakDecode(data, config)
	if err != nil {
		return nil, fmt.Errorf("could not decode bccsp type %w", err)
	}

	return config, nil
}

// EnhancedExactUnmarshal is intended to unmarshal a config file into a structure
// producing error when extraneous variables are introduced and supporting
// the time.Duration type
func (c *ConfigParser) EnhancedExactUnmarshal(output interface{}) error {
	oType := reflect.TypeOf(output)
	if oType.Kind() != reflect.Ptr {
		return fmt.Errorf("supplied output argument must be a pointer to a struct but is not pointer")
	}
	eType := oType.Elem()
	if eType.Kind() != reflect.Struct {
		return fmt.Errorf("supplied output argument must be a pointer to a struct, but it is pointer to something else")
	}

	baseKeys := c.config
	leafKeys := getKeysRecursively("", c.getFromEnv, baseKeys, eType)

	log.Printf("%+v", leafKeys)
	config := &mapstructure.DecoderConfig{
		ErrorUnused:      true,
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			bccspHook,
			mapstructure.StringToTimeDurationHookFunc(),
			customDecodeHook,
			byteSizeDecodeHook,
			stringFromFileDecodeHook,
			pemBlocksFromFileDecodeHook,
			kafkaVersionDecodeHook,
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(leafKeys)
}

// YamlStringToStructHook is a hook for viper(viper.Unmarshal(*,*, here)), it is able to parse a string of minified yaml into a slice of structs
func YamlStringToStructHook(m interface{}) func(rf reflect.Kind, rt reflect.Kind, data interface{}) (interface{}, error) {
	return func(rf reflect.Kind, rt reflect.Kind, data interface{}) (interface{}, error) {
		if rf != reflect.String || rt != reflect.Slice {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return m, nil
		}

		return m, yaml.UnmarshalStrict([]byte(raw), &m)
	}
}
