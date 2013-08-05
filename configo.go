package configo

import (
    "errors"
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "strings"
    "time"
)

/*
Configo maintains the set of valid configuration options as well as those read
in from a configuration file.
*/
type Configo struct {
    Path string
    Items map[string]*ConfigoItem
    configParsed bool
    flagsParsed bool
}

/*
ConfigoItem is a single configuration item registered to a Configo.
*/
type ConfigoItem struct {
    Name string
    Value interface{}
    Default interface{}
    Usage string
    IsFlag bool
    IsConfig bool
}

/*
New returns a newly initialized *Configo ready to have new ConfigoItems added
to it.
*/
func New(path string) *Configo {
    c := new(Configo)

    c.Path         = path
    c.Items        = make(map[string]*ConfigoItem, 100)
    c.configParsed = false
    c.flagsParsed  = false

    return c
}

/*
Get retrieves the value for a config item's key.
TODO 
*/
func (c *Configo) Get(key string) interface{} {
    return nil
}

/*
WriteDefaultConfig writes a config file to path which contains all of the
defined configuration items with their default values, including usage
comments.
TODO
*/
func (c *Configo) WriteDefaultConfig(path string) error {
    return errors.New("Not implemented!")
}

/*
Load reads in the config file at path and makes the key:value pairs available
to the program through the c.Items map.
*/
func (c *Configo) Load() (err error) {
    // Create the config file if it does not exist.
    if _, err = os.Stat(c.Path); err != nil {
        if os.IsNotExist(err) {
            c.configParsed = true

            if err = c.WriteDefaultConfig(c.Path); err != nil {
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

            c.Items[fields[0]].Value = fields[1]
        }
    }

    c.configParsed = true
    return
}

/*
Arg returns the i'th command-line argument. Arg(0) is the first remaining
argument after flags have been processed.
TODO
*/
func (c *Configo) Arg(i int) string {
    return ""
}

/*
Args returns the non-flag command-line arguments.
TODO
*/
func (c *Configo) Args() []string {
    return []string{}
}

/*
Bool defines a bool configuration option with specified name, default value,
and usage string.  The isFlag and isConfig parameters control whether the
option is valid on the command line and in the configuration file respectively.
TODO integrate flag pkg
*/
func (c *Configo) Bool(name string, value bool, usage string, isFlag, isConfig bool) {
    if _, exists := c.Items[name]; exists {
        panic(fmt.Sprintf("A configuration item named [%s] already exists!", name))
    }

    item := new(ConfigoItem)
    item.Name     = name
    item.Value    = value
    item.Default  = value
    item.Usage    = usage
    item.IsFlag   = isFlag
    item.IsConfig = isConfig

    c.Items[name] = item
}

/*
Duration defines a time.Duration flag with specified name, default value, and
usage string. The return value is the address of a time.Duration variable that
stores the value of the flag.
TODO fix doc
TODO integrate flag pkg
*/
func (c *Configo) Duration(name string, value time.Duration, usage string, isFlag, isConfig bool) {
    if _, exists := c.Items[name]; exists {
        panic(fmt.Sprintf("A configuration item named [%s] already exists!", name))
    }

    item := new(ConfigoItem)
    item.Name     = name
    item.Value    = value
    item.Default  = value
    item.Usage    = usage
    item.IsFlag   = isFlag
    item.IsConfig = isConfig

    c.Items[name] = item
}

/*
Float64 defines a float64 configuration item with specified name, default
value, and usage string.  The isFlag and isConfig parameters control whether
the option is valid on the command line and in the configuration file
respectively.
TODO integrate flag pkg
*/
func (c *Configo) Float64(name string, value float64, usage string, isFlag, isConfig bool) {
    if _, exists := c.Items[name]; exists {
        panic(fmt.Sprintf("A configuration item named [%s] already exists!", name))
    }

    item := new(ConfigoItem)
    item.Name     = name
    item.Value    = value
    item.Default  = value
    item.Usage    = usage
    item.IsFlag   = isFlag
    item.IsConfig = isConfig

    c.Items[name] = item
}

/*
Int defines an int flag with specified name, default value, and usage string.
The return value is the address of an int variable that stores the value of the
flag.
TODO fix doc
TODO integrate flag pkg
*/
func (c *Configo) Int(name string, value int, usage string, isFlag, isConfig bool) {
    if _, exists := c.Items[name]; exists {
        panic(fmt.Sprintf("A configuration item named [%s] already exists!", name))
    }

    item := new(ConfigoItem)
    item.Name     = name
    item.Value    = value
    item.Default  = value
    item.Usage    = usage
    item.IsFlag   = isFlag
    item.IsConfig = isConfig

    c.Items[name] = item
}

/*
Int64 defines an int64 flag with specified name, default value, and usage
string. The return value is the address of an int64 variable that stores the
value of the flag.
TODO fix doc
TODO integrate flag pkg
*/
func (c *Configo) Int64(name string, value int64, usage string, isFlag, isConfig bool) {
    if _, exists := c.Items[name]; exists {
        panic(fmt.Sprintf("A configuration item named [%s] already exists!", name))
    }

    item := new(ConfigoItem)
    item.Name     = name
    item.Value    = value
    item.Default  = value
    item.Usage    = usage
    item.IsFlag   = isFlag
    item.IsConfig = isConfig

    c.Items[name] = item
}

/*
NArg is the number of arguments remaining after flags have been processed.
TODO
*/
func (c *Configo) NArg() int {
    return 0
}

/*
NFlag returns the number of command-line flags that have been set.
TODO
*/
func (c *Configo) NFlag() int {
    return 0
}

/*
Parse parses the command-line flags from os.Args[1:]. Must be called after all
flags are defined and before flags are accessed by the program.
TODO
*/
func (c *Configo) Parse() {
}

/*
Parsed returns true if the command-line flags have been parsed.
TODO
*/
func (c *Configo) Parsed() bool {
    return false
}

/*
PrintDefaults prints to standard error the default values of all defined
command-line flags.
TODO
*/
func (c *Configo) PrintDefaults() {
}

/*
Set sets the value of the named command-line flag.
TODO
*/
func (c *Configo) Set(name, value string) error {
    return errors.New("Not implemented.")
}

/*
String defines a string flag with specified name, default value, and usage
string. The return value is the address of a string variable that stores the
value of the flag.
TODO fix doc
TODO integrate flag pkg
*/
func (c *Configo) String(name string, value string, usage string, isFlag, isConfig bool) {
    if _, exists := c.Items[name]; exists {
        panic(fmt.Sprintf("A configuration item named [%s] already exists!", name))
    }

    item := new(ConfigoItem)
    item.Name     = name
    item.Value    = value
    item.Default  = value
    item.Usage    = usage
    item.IsFlag   = isFlag
    item.IsConfig = isConfig

    c.Items[name] = item
}

/*
Uint defines a uint flag with specified name, default value, and usage string.
The return value is the address of a uint variable that stores the value of the
flag.
TODO fix doc
TODO integrate flag pkg
*/
func (c *Configo) Uint(name string, value uint, usage string, isFlag, isConfig bool) {
    if _, exists := c.Items[name]; exists {
        panic(fmt.Sprintf("A configuration item named [%s] already exists!", name))
    }

    item := new(ConfigoItem)
    item.Name     = name
    item.Value    = value
    item.Default  = value
    item.Usage    = usage
    item.IsFlag   = isFlag
    item.IsConfig = isConfig

    c.Items[name] = item
}

/*
Uint64 defines a uint64 flag with specified name, default value, and usage
string. The return value is the address of a uint64 variable that stores the
value of the flag.
TODO fix doc
TODO integrate flag pkg
*/
func (c *Configo) Uint64(name string, value uint64, usage string, isFlag, isConfig bool) {
    if _, exists := c.Items[name]; exists {
        panic(fmt.Sprintf("A configuration item named [%s] already exists!", name))
    }

    item := new(ConfigoItem)
    item.Name     = name
    item.Value    = value
    item.Default  = value
    item.Usage    = usage
    item.IsFlag   = isFlag
    item.IsConfig = isConfig

    c.Items[name] = item
}

/*
Var defines a flag with the specified name and usage string. The type and value
of the flag are represented by the first argument, of type Value, which
typically holds a user-defined implementation of Value. For instance, the
caller could create a flag that turns a comma-separated string into a slice of
strings by giving the slice the methods of Value; in particular, Set would
decompose the comma-separated string into the slice.
TODO fix doc
TODO integrate flag pkg
*/
func (c *Configo) Var(value flag.Value, name string, usage string, isFlag, isConfig bool) {
    if _, exists := c.Items[name]; exists {
        panic(fmt.Sprintf("A configuration item named [%s] already exists!", name))
    }

    item := new(ConfigoItem)
    item.Name     = name
    item.Value    = value
    item.Default  = value
    item.Usage    = usage
    item.IsFlag   = isFlag
    item.IsConfig = isConfig

    c.Items[name] = item
}

/*
Visit visits the command-line flags in lexicographical order, calling fn for
each. It visits only those flags that have been set.
TODO
*/
func (c *Configo) Visit(fn func(*flag.Flag)) {
}

/*
VisitAll visits the command-line flags in lexicographical order, calling fn for
each. It visits all flags, even those not set.
TODO
*/
func (c *Configo) VisitAll(fn func(*flag.Flag)) {
}

