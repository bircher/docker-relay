package main

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

// Config is parsed from the config file.
type config struct {
	// docker command config
	Image string
	Cmd   []string
	Path  string
	Exec  []string

	// docker exec config
	User    string   `option:"user"`
	Workdir string   `option:"workdir"`
	Env     []string `option:"env"`
	Detach  bool     `option:"detach"`

	// docker run config, also uses exec config
	Name   string   `option:"name"`
	Link   []string `option:"link"`
	Volume []string `option:"volume"`
	Rm     bool     `option:"rm"`
	Device []string `option:"device"`
}

// options returns the options to pass to docker
func (c *config) options() []string {

	list := []string{}
	v := reflect.ValueOf(c)
	// v is a pointer, get the element
	v = v.Elem()

	// Loop through all the fields.
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		// Get the option tag and split it by comma, in case we add something else.
		opts := strings.Split(v.Type().Field(i).Tag.Get("option"), ",")
		if len(opts) == 0 || opts[0] == "" {
			// If there are no options or the option is empty, it is not an option.
			continue
		}
		opt := "--" + opts[0]

		switch f.Kind() {
		case reflect.String:
			if len(f.String()) > 0 {
				// If it is a string and the field is not empty, add the option.
				list = append(list, opt, f.String())
			}
		case reflect.Slice:
			li, ok := f.Interface().([]string)
			if !ok {
				panic("value not a []string")
			}
			// For slices add all the elements of the slice as options.
			for _, el := range li {
				if len(el) > 0 {
					list = append(list, opt, el)
				}
			}
		case reflect.Bool:
			if f.Bool() {
				// For a boolean value add the option if it is true.
				list = append(list, opt)
			}
		}

	}

	return list
}

// replaceStrings searches for strings in some configuration keys and replaces them with what the function returns for them.
func (c *config) replaceStrings(r *regexp.Regexp, trans func(string, string) string) {

	// The function to fix the string used by the reflection processing.
	fix := func(item string, key string) string {
		fixed := item
		// Find all regex matches.
		for _, s := range r.FindAllString(item, -1) {
			// Replace the matches with the translation function.
			fixed = strings.Replace(fixed, s, trans(s, key), 1)
		}
		return fixed
	}

	v := reflect.ValueOf(c)
	// v is a pointer, get the element
	v = v.Elem()

	// Loop through all the fields.
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		n := v.Type().Field(i).Name
		switch f.Kind() {
		case reflect.String:
			if len(f.String()) > 0 {
				// Fix the string with our fix function.
				f.SetString(fix(f.String(), n))
			}
		case reflect.Slice:
			li, ok := f.Interface().([]string)
			if !ok {
				panic("value not a []string")
			}

			for j, el := range li {
				if len(el) > 0 {
					// Fix each element of the slice with our fix function.
					f.Index(j).SetString(fix(el, n))
				}
			}
		}
	}
}

// getConf returns the configuration for the program
func getConf(cfg *viper.Viper, run bool) (*config, error) {
	var c *config

	if err := cfg.Unmarshal(&c); err != nil {
		return nil, err
	}

	if run {
		// Set default values for when running an image
		c.Rm = true

		rcfg := cfg.Sub("run")
		if rcfg != nil {
			if err := rcfg.Unmarshal(&c); err != nil {
				return c, err
			}
		}
	}

	return c, nil
}
