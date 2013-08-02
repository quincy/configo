package configo

import (
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "strings"
)

/* Configo maintains the set of valid configuration options as well as those
read in from a configuration file. */
type Configo struct {
    Path string
    Items map[string]string
    Options map[string]ConfigoItem
    loaded bool
}

/* ConfigoItem is a single configuration item registered to a Configo. */
type ConfigoItem struct {
    Key string
    Value string
    Default string
    Help string
}

/* New returns a newly initialized *Configo ready to have new ConfigoItems
added to it. */
func New(path string) *Configo {
    c := new(Configo)

    c.Path    = path
    c.Items   = make(map[string]string, 100)
    c.Options = make(map[string]ConfigoItem, 100)
    c.loaded  = false

    return c
}

/* Add adds a new configuration item to this Configo. */
func (c *Configo) Add(key, value, defaultValue, help string) (err error) {
    if _, exists := c.Options[key]; exists {
        err = errors.New(fmt.Sprintf("A config item with key [%s] has already been added to this Configo.", key))
    } else {
        c.Options[key] = ConfigoItem{
            Key:     key,
            Value:   value,
            Default: defaultValue,
            Help:    help }
    }

    return
}

/* Get retrieves the value for a config item's key. */
func (c *Configo) Get(key string) string {
    if c.loaded {
        if item, ok := c.Items[key]; ok {
            return item.Value
        }

        return item.Default
    }

    panic("Attempt to get config item before loading configuration values from file.")
}

/* Load reads in the config file at path and makes the key:value pairs
available to the program through the c.Items map. */
func (c *Configo) Load() (err error) {
    // Create the config file if it does not exist.
    if _, err = os.Stat(c.Path); err != nil {
        if os.IsNotExist(err) {
            c.Items  = c.getDefaultValues()
            c.loaded = true

            if err = c.WriteConfig(c.Path); err != nil {
                return
            }
        }
        return
    }

    content, err := ioutil.ReadFile(c.Path)
    if err != nil {
        return
    }

    for i, line := range strings.Split(string(content), "\n") {
        line = strings.TrimSpace(line)

        if len(line) > 0 && !strings.HasPrefix(line, "#") {
            fields := strings.SplitN(line, ":", 2)
            if len(fields) != 2 {
                errors.New(fmt.Sprintf("Invalid key:value pair in conifiguration file %s on line %d.\n", c.Path, i))
            }

            if err := c.Valid(fields[0], fields[1]); err != nil {
                return
            }

            c.Items[fields[0]] = fields[1]
        }
    }

    c.loaded = true
    return
}

/* Valid checks if a given key/value pair is a valid configuration option.
It returns an error if key is not a valid configuration item or if value is
not a valid value for the key. */
func (c *Configo) Valid(key, value string) error {
    if option, ok := c.Options[key]; ok {
        if err := option.Validate(value); err != nil {
            return err
        }

        return nil
    }

    return errors.New(fmt.Sprintf("Invalid key [%s] in configuration file.\n", key))
}

/* Validate verifies the given value is okay for this ConfigoItem. */
func (ci *ConfigoItem) Validate(v string) error {
    // TODO Implement value validation.
    return nil
    //return errors.New(fmt.Sprintf("Invalid value [%s] for key [%s] in configuration file.\n", v, ci.Key))
}

/* writeDefaultConfig writes a new configuration file to path, containing all
of the configuration items with default values and their help text as
comments. */
func (c *Configo) writeDefaultConfig(path string) (err error) {
    // TODO
    c.loaded = true
    return
}

