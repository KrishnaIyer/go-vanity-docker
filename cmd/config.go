// Copyright © 2020 Krishna Iyer Easwaran
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	commaSeparator = ","
	colonSeparator = ":"
)

// Manager is the configuration manager.
type Manager struct {
	name  string
	flags *pflag.FlagSet
	viper *viper.Viper
}

// New returns a new initialized manager with the given config.
func New(name, prefix string) *Manager {
	viper := viper.New()
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()
	viper.SetConfigName(name)
	viper.SetEnvPrefix(strings.ToUpper(prefix))

	return &Manager{
		name:  name,
		flags: pflag.NewFlagSet(name, pflag.ExitOnError),
		viper: viper,
	}
}

// InitFlags initializes the flagset with the provided config.
func (mgr *Manager) InitFlags(cfg interface{}) error {
	rootStruct := reflect.TypeOf(cfg)

	if rootStruct.Kind() != reflect.Struct {
		panic("configuration is not a struct")
	}

	mgr.parseStructToFlags("", rootStruct)

	err := mgr.viper.BindPFlags(mgr.flags)
	if err != nil {
		panic(err)
	}

	return nil
}

// Flags returns pflag.FlagSet.
func (mgr *Manager) Flags() *pflag.FlagSet {
	return mgr.flags
}

// Unmarshal unmarshals the read config into the provided struct.
func (mgr *Manager) Unmarshal(config interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "name",
		Result:  config,
	})
	if err != nil {
		return err
	}

	decoder.Decode(mgr.viper.AllSettings())
	return nil
}

// Viper returns viper.
func (mgr *Manager) Viper() *viper.Viper {
	return mgr.viper
}

// parseStructToFlags parses a struct and returns a flagset.
// It panics if there are parsing errors.
func (mgr *Manager) parseStructToFlags(prefix string, strT reflect.Type) {
	for i := 0; i < strT.NumField(); i++ {
		field := strT.Field(i)
		name := field.Tag.Get("name")
		kind := field.Type.Kind()
		if (name == "" || name == "-") && kind != reflect.Struct {
			continue
		}

		desc := field.Tag.Get("description")
		short := field.Tag.Get("short")

		if prefix != "" {
			name = prefix + "." + name
		}

		switch kind {
		case reflect.String:
			mgr.flags.StringP(name, short, "", desc)
		case reflect.Bool:
			mgr.flags.BoolP(name, short, false, desc)
		case reflect.Int:
			mgr.flags.IntP(name, short, 0, desc)
		case reflect.Struct:
			// This allows for recursion
			mgr.parseStructToFlags(name, field.Type)
		default:
			panic(fmt.Errorf("Unknown type in config: %v", kind))
		}
	}
}
