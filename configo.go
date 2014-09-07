// The MIT License (MIT)
//
// Copyright (c) 2013 Quincy Bowers
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

/*
Package configo extends the standard flag package to include configuration file
parsing and management.

Usage:

Usage is almost identical to the flag package: declare flags, parse flags, use
flags.  Command line flags override values parsed from a configuration file.

Configuration files consist of lines of key/value pairs, delimited by '='.  The
delimiter can be changed if needed by setting configo.SetDelimiter().  Blank lines
and lines where the first non-whitespace character is '#' are ignored.
Trailing comments are not allowed, however.
*/
package configo

import (
    "errors"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "os/user"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "time"
)

// ConfigoSet maintains the set of valid configuration options as well as those
// read in from a configuration file.
type ConfigoSet struct {
    Usage func()

    name          string
    parsed        bool
    actual        map[string]*Configo
    formal        map[string]*Configo
    exitOnError   flag.ErrorHandling
    errorHandling flag.ErrorHandling
    output        io.Writer
    path          string
    delimiter     string
}

// Configo is a single configuration item registered to a ConfigoSet.
type Configo struct {
    Name         string
    Usage        string
    Value        flag.Value
    DefaultValue string
    IsFlag       bool
    IsConfig     bool
}

// -- bool Value
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
    *p = val
    return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
    v, err := strconv.ParseBool(s)
    *b = boolValue(v)
    return err
}

func (b *boolValue) String() string { return fmt.Sprintf("%v", *b) }

func (b *boolValue) IsBoolFlag() bool { return true }

// optional interface to indicate boolean flags that can be
// supplied without "=value" text
type boolFlag interface {
    flag.Value
    IsBoolFlag() bool
}

// -- int Value
type intValue int

func newIntValue(val int, p *int) *intValue {
    *p = val
    return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
    v, err := strconv.ParseInt(s, 0, 64)
    *i = intValue(v)
    return err
}

func (i *intValue) String() string { return fmt.Sprintf("%v", *i) }

// -- int64 Value
type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
    *p = val
    return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
    v, err := strconv.ParseInt(s, 0, 64)
    *i = int64Value(v)
    return err
}

func (i *int64Value) String() string { return fmt.Sprintf("%v", *i) }

// -- uint Value
type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
    *p = val
    return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
    v, err := strconv.ParseUint(s, 0, 64)
    *i = uintValue(v)
    return err
}

func (i *uintValue) String() string { return fmt.Sprintf("%v", *i) }

// -- uint64 Value
type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
    *p = val
    return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
    v, err := strconv.ParseUint(s, 0, 64)
    *i = uint64Value(v)
    return err
}

func (i *uint64Value) String() string { return fmt.Sprintf("%v", *i) }

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
    *p = val
    return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
    *s = stringValue(val)
    return nil
}

func (s *stringValue) String() string { return fmt.Sprintf("%s", *s) }

// -- float64 Value
type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
    *p = val
    return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
    v, err := strconv.ParseFloat(s, 64)
    *f = float64Value(v)
    return err
}

func (f *float64Value) String() string { return fmt.Sprintf("%v", *f) }

// -- time.Duration Value
type durationValue time.Duration

func newDurationValue(val time.Duration, p *time.Duration) *durationValue {
    *p = val
    return (*durationValue)(p)
}

func (d *durationValue) Set(s string) error {
    v, err := time.ParseDuration(s)
    *d = durationValue(v)
    return err
}

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

// The default set of configuration options.
var baseProgName string = filepath.Base(os.Args[0])
var configuration = NewConfigoSet(baseProgName, flag.ExitOnError, DefaultConfigPath())

// NewConfigoSet returns a new, empty configuration set with the specified name
// and error handling property.
func NewConfigoSet(name string, errorHandling flag.ErrorHandling, path string) *ConfigoSet {
    c := &ConfigoSet{
        name:          name,
        errorHandling: errorHandling,
        delimiter:     "=",
        path:          path,
    }
    return c
}

// defaultConfigPath returns the default configuration file path which is
// either in the current user's home directory, if there is a current user, or
// in the current working directory.  The name of the config file will be the
// standard unix naming convention "." + {ProgramName} + "rc".
func DefaultConfigPath() string {
    usr, err := user.Current()
    if err != nil {
        return fmt.Sprintf(".%src", baseProgName)
    }
    return fmt.Sprintf(".%src", filepath.Join(usr.HomeDir, baseProgName))
}

// SetPath sets the path to the configuration file.
func SetPath(path string) {
    configuration.path = path
}

// WriteDefaultConfig writes a config file to c.path which contains all of the
// defined configuration items with their default values, including usage
// comments.
func (c *ConfigoSet) WriteDefaultConfig(path string) (err error) {
    fmt.Fprintln(c.out(), "Writing a default configuration file to", path)

    origOut := c.output
    c.output, err = os.Create(c.path)
    if err != nil {
        return
    }

    fmt.Fprintf(c.out(), "# Default config file for %s\n", c.name)
    fmt.Fprintf(c.out(), "# Written on %s\n\n", time.Now().Format(time.RFC822Z))

    c.VisitAll(func(config *Configo) {
        if config.IsConfig {
            format := "# %s\n%s%s%s\n\n"
            fmt.Fprintf(c.out(), format, config.Usage, config.Name, c.delimiter, config.DefaultValue)
        }
    })

    c.output = origOut
    return
}

// Arg returns the i'th command-line argument. Arg(0) is the first remaining
// argument after flags have been processed.
func (c *ConfigoSet) Arg(i int) string {
    return flag.Arg(i)
}

// Args returns the non-flag command-line arguments.
func (c *ConfigoSet) Args() []string {
    return flag.Args()
}

// -- User functions for registering bool flags

// BoolVar defines a bool config item with specified name, default value, and
// usage string.  The argument p points to a bool variable in which to store
// the value of the flag.
//
// This item can be specified on the command line and in the configuration
// file.
func (c *ConfigoSet) BoolVar(p *bool, name string, value bool, usage string) {
    isFlag := true
    isConfig := true
    c.Var(newBoolValue(value, p), name, usage, isFlag, isConfig)
    flag.BoolVar(p, name, value, usage)
}

// BoolConfigVar defines a bool config item with specified name, default value,
// and usage string.  The argument p points to a bool variable in which to
// store the value of the flag.
//
// This item can only be specified in the configuration file.
func (c *ConfigoSet) BoolConfigVar(p *bool, name string, value bool, usage string) {
    isFlag := false
    isConfig := true
    c.Var(newBoolValue(value, p), name, usage, isFlag, isConfig)
}

// BoolFlagVar defines a bool command line flag item with specified name,
// default value, and usage string.  The argument p points to a bool variable
// in which to store the value of the flag.
//
// This item can only be specified on the command line.
func (c *ConfigoSet) BoolFlagVar(p *bool, name string, value bool, usage string) {
    isFlag := true
    isConfig := false
    c.Var(newBoolValue(value, p), name, usage, isFlag, isConfig)
    flag.BoolVar(p, name, value, usage)
}

// BoolVar defines a bool config item with specified name, default value, and
// usage string.  The argument p points to a bool variable in which to store
// the value of the flag.
//
// This item can be specified on the command line and in the configuration
// file.
func BoolVar(p *bool, name string, value bool, usage string) {
    isFlag := true
    isConfig := true
    configuration.Var(newBoolValue(value, p), name, usage, isFlag, isConfig)
    flag.BoolVar(p, name, value, usage)
}

// BoolConfigVar defines a bool config item with specified name, default value, and
// usage string.  The argument p points to a bool variable in which to store
// the value of the flag.
//
// This item can only be specified in the configuration file.
func BoolConfigVar(p *bool, name string, value bool, usage string) {
    isFlag := false
    isConfig := true
    configuration.Var(newBoolValue(value, p), name, usage, isFlag, isConfig)
}

// BoolFlagVar defines a bool config item with specified name, default value, and
// usage string.  The argument p points to a bool variable in which to store
// the value of the flag.
//
// This item can only be specified on the command line.
func BoolFlagVar(p *bool, name string, value bool, usage string) {
    isFlag := true
    isConfig := false
    configuration.Var(newBoolValue(value, p), name, usage, isFlag, isConfig)
    flag.BoolVar(p, name, value, usage)
}

// Bool defines a bool configuration option with specified name, default value,
// and usage string.  The isFlag and isConfig parameters control whether the
// option is valid on the command line and in the configuration file respectively.
//
// This item can be specified on the command line and in the configuration
// file.
func (c *ConfigoSet) Bool(name string, value bool, usage string) *bool {
    p := new(bool)
    c.BoolVar(p, name, value, usage)
    return p
}

// BoolFlag defines a bool configuration option with specified name, default value,
// and usage string.
//
// This item can only be specified on the command line.
func (c *ConfigoSet) BoolFlag(name string, value bool, usage string) *bool {
    p := new(bool)
    c.BoolFlagVar(p, name, value, usage)
    return p
}

// BoolConfig defines a bool configuration option with specified name, default
// value, and usage string.
//
// This item can only be specified in the configuration file.
func (c *ConfigoSet) BoolConfig(name string, value bool, usage string) *bool {
    p := new(bool)
    c.BoolConfigVar(p, name, value, usage)
    return p
}

// Bool defines a bool config item with specified name, default value, and
// usage string.  The return value is the address of a bool variable that
// stores the value of the config item.
//
// This item can be specified on the command line and in the configuration
// file.
func Bool(name string, value bool, usage string) *bool {
    return configuration.Bool(name, value, usage)
}

// BoolFlag defines a bool config item with specified name, default value, and
// usage string.  The return value is the address of a bool variable that
// stores the value of the config item.
func BoolFlag(name string, value bool, usage string) *bool {
    return configuration.BoolFlag(name, value, usage)
}

// BoolConfig defines a bool config item with specified name, default value, and
// usage string.  The return value is the address of a bool variable that
// stores the value of the config item.
func BoolConfig(name string, value bool, usage string) *bool {
    return configuration.BoolConfig(name, value, usage)
}

// -- User functions for registering Int flags

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (c *ConfigoSet) IntVar(p *int, name string, value int, usage string) {
    isFlag := true
    isConfig := true
    c.Var(newIntValue(value, p), name, usage, isFlag, isConfig)
    flag.IntVar(p, name, value, usage)
}

// IntFlagVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (c *ConfigoSet) IntFlagVar(p *int, name string, value int, usage string) {
    isFlag := true
    isConfig := false
    c.Var(newIntValue(value, p), name, usage, isFlag, isConfig)
    flag.IntVar(p, name, value, usage)
}

// IntConfigVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (c *ConfigoSet) IntConfigVar(p *int, name string, value int, usage string) {
    isFlag := false
    isConfig := true
    c.Var(newIntValue(value, p), name, usage, isFlag, isConfig)
    flag.IntVar(p, name, value, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntVar(p *int, name string, value int, usage string) {
    isFlag := true
    isConfig := true
    configuration.Var(newIntValue(value, p), name, usage, isFlag, isConfig)
    flag.IntVar(p, name, value, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntConfigVar(p *int, name string, value int, usage string) {
    isFlag := false
    isConfig := true
    configuration.Var(newIntValue(value, p), name, usage, isFlag, isConfig)
    flag.IntVar(p, name, value, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntFlagVar(p *int, name string, value int, usage string) {
    isFlag := true
    isConfig := false
    configuration.Var(newIntValue(value, p), name, usage, isFlag, isConfig)
    flag.IntVar(p, name, value, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (c *ConfigoSet) Int(name string, value int, usage string) *int {
    p := new(int)
    c.IntVar(p, name, value, usage)
    return p
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (c *ConfigoSet) IntConfig(name string, value int, usage string) *int {
    p := new(int)
    c.IntConfigVar(p, name, value, usage)
    return p
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (c *ConfigoSet) IntFlag(name string, value int, usage string) *int {
    p := new(int)
    c.IntFlagVar(p, name, value, usage)
    return p
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func Int(name string, value int, usage string) *int {
    return configuration.Int(name, value, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func IntConfig(name string, value int, usage string) *int {
    return configuration.IntConfig(name, value, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func IntFlag(name string, value int, usage string) *int {
    return configuration.IntFlag(name, value, usage)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func (c *ConfigoSet) Int64Var(p *int64, name string, value int64, usage string) {
    isFlag := true
    isConfig := true
    c.Var(newInt64Value(value, p), name, usage, isFlag, isConfig)
    flag.Int64Var(p, name, value, usage)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func (c *ConfigoSet) Int64FlagVar(p *int64, name string, value int64, usage string) {
    isFlag := true
    isConfig := false
    c.Var(newInt64Value(value, p), name, usage, isFlag, isConfig)
    flag.Int64Var(p, name, value, usage)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func (c *ConfigoSet) Int64ConfigVar(p *int64, name string, value int64, usage string) {
    isFlag := false
    isConfig := true
    c.Var(newInt64Value(value, p), name, usage, isFlag, isConfig)
    flag.Int64Var(p, name, value, usage)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func Int64Var(p *int64, name string, value int64, usage string) {
    isFlag := true
    isConfig := false
    configuration.Var(newInt64Value(value, p), name, usage, isFlag, isConfig)
    flag.Int64Var(p, name, value, usage)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func Int64ConfigVar(p *int64, name string, value int64, usage string) {
    isFlag := false
    isConfig := true
    configuration.Var(newInt64Value(value, p), name, usage, isFlag, isConfig)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func Int64FlagVar(p *int64, name string, value int64, usage string) {
    isFlag := true
    isConfig := false
    configuration.Var(newInt64Value(value, p), name, usage, isFlag, isConfig)
    flag.Int64Var(p, name, value, usage)
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func (c *ConfigoSet) Int64(name string, value int64, usage string) *int64 {
    p := new(int64)
    c.Int64Var(p, name, value, usage)
    return p
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func (c *ConfigoSet) Int64Flag(name string, value int64, usage string) *int64 {
    p := new(int64)
    c.Int64FlagVar(p, name, value, usage)
    return p
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func (c *ConfigoSet) Int64Config(name string, value int64, usage string) *int64 {
    p := new(int64)
    c.Int64ConfigVar(p, name, value, usage)
    return p
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func Int64(name string, value int64, usage string) *int64 {
    return configuration.Int64(name, value, usage)
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func Int64Flag(name string, value int64, usage string) *int64 {
    return configuration.Int64Flag(name, value, usage)
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func Int64Config(name string, value int64, usage string) *int64 {
    return configuration.Int64Config(name, value, usage)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func (c *ConfigoSet) UintVar(p *uint, name string, value uint, usage string) {
    isFlag := true
    isConfig := true
    c.Var(newUintValue(value, p), name, usage, isFlag, isConfig)
    flag.UintVar(p, name, value, usage)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func (c *ConfigoSet) UintFlagVar(p *uint, name string, value uint, usage string) {
    isFlag := true
    isConfig := false
    c.Var(newUintValue(value, p), name, usage, isFlag, isConfig)
    flag.UintVar(p, name, value, usage)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func (c *ConfigoSet) UintConfigVar(p *uint, name string, value uint, usage string) {
    isFlag := true
    isConfig := false
    c.Var(newUintValue(value, p), name, usage, isFlag, isConfig)
    flag.UintVar(p, name, value, usage)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint  variable in which to store the value of the flag.
func UintVar(p *uint, name string, value uint, usage string) {
    isFlag := true
    isConfig := true
    configuration.Var(newUintValue(value, p), name, usage, isFlag, isConfig)
    flag.UintVar(p, name, value, usage)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint  variable in which to store the value of the flag.
func UintFlagVar(p *uint, name string, value uint, usage string) {
    isFlag := true
    isConfig := false
    configuration.Var(newUintValue(value, p), name, usage, isFlag, isConfig)
    flag.UintVar(p, name, value, usage)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint  variable in which to store the value of the flag.
func UintConfigVar(p *uint, name string, value uint, usage string) {
    isFlag := false
    isConfig := true
    configuration.Var(newUintValue(value, p), name, usage, isFlag, isConfig)
    flag.UintVar(p, name, value, usage)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint  variable that stores the value of the flag.
func (c *ConfigoSet) Uint(name string, value uint, usage string) *uint {
    p := new(uint)
    c.UintVar(p, name, value, usage)
    return p
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint  variable that stores the value of the flag.
func (c *ConfigoSet) UintFlag(name string, value uint, usage string) *uint {
    p := new(uint)
    c.UintFlagVar(p, name, value, usage)
    return p
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint  variable that stores the value of the flag.
func (c *ConfigoSet) UintConfig(name string, value uint, usage string) *uint {
    p := new(uint)
    c.UintVar(p, name, value, usage)
    return p
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint  variable that stores the value of the flag.
func Uint(name string, value uint, usage string) *uint {
    return configuration.Uint(name, value, usage)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint  variable that stores the value of the flag.
func UintFlag(name string, value uint, usage string) *uint {
    return configuration.UintFlag(name, value, usage)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint  variable that stores the value of the flag.
func UintConfig(name string, value uint, usage string) *uint {
    return configuration.UintConfig(name, value, usage)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func (c *ConfigoSet) Uint64Var(p *uint64, name string, value uint64, usage string) {
    isFlag := true
    isConfig := true
    c.Var(newUint64Value(value, p), name, usage, isFlag, isConfig)
    flag.Uint64Var(p, name, value, usage)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func (c *ConfigoSet) Uint64FlagVar(p *uint64, name string, value uint64, usage string) {
    isFlag := true
    isConfig := false
    c.Var(newUint64Value(value, p), name, usage, isFlag, isConfig)
    flag.Uint64Var(p, name, value, usage)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func (c *ConfigoSet) Uint64ConfigVar(p *uint64, name string, value uint64, usage string) {
    isFlag := false
    isConfig := true
    c.Var(newUint64Value(value, p), name, usage, isFlag, isConfig)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func Uint64Var(p *uint64, name string, value uint64, usage string) {
    isFlag := true
    isConfig := true
    configuration.Var(newUint64Value(value, p), name, usage, isFlag, isConfig)
    flag.Uint64Var(p, name, value, usage)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func Uint64FlagVar(p *uint64, name string, value uint64, usage string) {
    isFlag := true
    isConfig := false
    configuration.Var(newUint64Value(value, p), name, usage, isFlag, isConfig)
    flag.Uint64Var(p, name, value, usage)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func Uint64ConfigVar(p *uint64, name string, value uint64, usage string) {
    isFlag := true
    isConfig := true
    configuration.Var(newUint64Value(value, p), name, usage, isFlag, isConfig)
    flag.Uint64Var(p, name, value, usage)
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func (c *ConfigoSet) Uint64(name string, value uint64, usage string) *uint64 {
    p := new(uint64)
    c.Uint64Var(p, name, value, usage)
    return p
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func (c *ConfigoSet) Uint64Flag(name string, value uint64, usage string) *uint64 {
    p := new(uint64)
    c.Uint64FlagVar(p, name, value, usage)
    return p
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func (c *ConfigoSet) Uint64Config(name string, value uint64, usage string) *uint64 {
    p := new(uint64)
    c.Uint64ConfigVar(p, name, value, usage)
    return p
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func Uint64(name string, value uint64, usage string) *uint64 {
    return configuration.Uint64(name, value, usage)
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func Uint64Flag(name string, value uint64, usage string) *uint64 {
    return configuration.Uint64Flag(name, value, usage)
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func Uint64Config(name string, value uint64, usage string) *uint64 {
    return configuration.Uint64Config(name, value, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (c *ConfigoSet) StringVar(p *string, name string, value string, usage string) {
    isFlag := true
    isConfig := true
    c.Var(newStringValue(value, p), name, usage, isFlag, isConfig)
    flag.StringVar(p, name, value, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (c *ConfigoSet) StringFlagVar(p *string, name string, value string, usage string) {
    isFlag := true
    isConfig := false
    c.Var(newStringValue(value, p), name, usage, isFlag, isConfig)
    flag.StringVar(p, name, value, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (c *ConfigoSet) StringConfigVar(p *string, name string, value string, usage string) {
    isFlag := false
    isConfig := true
    c.Var(newStringValue(value, p), name, usage, isFlag, isConfig)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name string, value string, usage string) {
    isFlag := true
    isConfig := true
    configuration.Var(newStringValue(value, p), name, usage, isFlag, isConfig)
    flag.StringVar(p, name, value, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringFlagVar(p *string, name string, value string, usage string) {
    isFlag := true
    isConfig := false
    configuration.Var(newStringValue(value, p), name, usage, isFlag, isConfig)
    flag.StringVar(p, name, value, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringConfigVar(p *string, name string, value string, usage string) {
    isFlag := false
    isConfig := true
    configuration.Var(newStringValue(value, p), name, usage, isFlag, isConfig)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (c *ConfigoSet) String(name string, value string, usage string) *string {
    p := new(string)
    c.StringVar(p, name, value, usage)
    return p
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (c *ConfigoSet) StringFlag(name string, value string, usage string) *string {
    p := new(string)
    c.StringFlagVar(p, name, value, usage)
    return p
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (c *ConfigoSet) StringConfig(name string, value string, usage string) *string {
    p := new(string)
    c.StringFlagVar(p, name, value, usage)
    return p
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name string, value string, usage string) *string {
    return configuration.String(name, value, usage)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func StringFlag(name string, value string, usage string) *string {
    return configuration.StringFlag(name, value, usage)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func StringConfig(name string, value string, usage string) *string {
    return configuration.StringConfig(name, value, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (c *ConfigoSet) Float64Var(p *float64, name string, value float64, usage string) {
    isFlag := true
    isConfig := true
    c.Var(newFloat64Value(value, p), name, usage, isFlag, isConfig)
    flag.Float64Var(p, name, value, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (c *ConfigoSet) Float64FlagVar(p *float64, name string, value float64, usage string) {
    isFlag := true
    isConfig := false
    c.Var(newFloat64Value(value, p), name, usage, isFlag, isConfig)
    flag.Float64Var(p, name, value, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (c *ConfigoSet) Float64ConfigVar(p *float64, name string, value float64, usage string) {
    isFlag := false
    isConfig := true
    c.Var(newFloat64Value(value, p), name, usage, isFlag, isConfig)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64Var(p *float64, name string, value float64, usage string) {
    isFlag := true
    isConfig := true
    configuration.Var(newFloat64Value(value, p), name, usage, isFlag, isConfig)
    flag.Float64Var(p, name, value, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64FlagVar(p *float64, name string, value float64, usage string) {
    isFlag := true
    isConfig := false
    configuration.Var(newFloat64Value(value, p), name, usage, isFlag, isConfig)
    flag.Float64Var(p, name, value, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64ConfigVar(p *float64, name string, value float64, usage string) {
    isFlag := false
    isConfig := true
    configuration.Var(newFloat64Value(value, p), name, usage, isFlag, isConfig)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func (c *ConfigoSet) Float64(name string, value float64, usage string) *float64 {
    p := new(float64)
    c.Float64Var(p, name, value, usage)
    return p
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func (c *ConfigoSet) Float64Flag(name string, value float64, usage string) *float64 {
    p := new(float64)
    c.Float64FlagVar(p, name, value, usage)
    return p
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func (c *ConfigoSet) Float64Config(name string, value float64, usage string) *float64 {
    p := new(float64)
    c.Float64ConfigVar(p, name, value, usage)
    return p
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float64(name string, value float64, usage string) *float64 {
    return configuration.Float64(name, value, usage)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float64Flag(name string, value float64, usage string) *float64 {
    return configuration.Float64Flag(name, value, usage)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float64Config(name string, value float64, usage string) *float64 {
    return configuration.Float64Config(name, value, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func (c *ConfigoSet) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
    isFlag := true
    isConfig := true
    c.Var(newDurationValue(value, p), name, usage, isFlag, isConfig)
    flag.DurationVar(p, name, value, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func (c *ConfigoSet) DurationFlagVar(p *time.Duration, name string, value time.Duration, usage string) {
    isFlag := true
    isConfig := false
    c.Var(newDurationValue(value, p), name, usage, isFlag, isConfig)
    flag.DurationVar(p, name, value, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func (c *ConfigoSet) DurationConfigVar(p *time.Duration, name string, value time.Duration, usage string) {
    isFlag := false
    isConfig := true
    c.Var(newDurationValue(value, p), name, usage, isFlag, isConfig)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
    isFlag := true
    isConfig := true
    configuration.Var(newDurationValue(value, p), name, usage, isFlag, isConfig)
    flag.DurationVar(p, name, value, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func DurationFlagVar(p *time.Duration, name string, value time.Duration, usage string) {
    isFlag := true
    isConfig := false
    configuration.Var(newDurationValue(value, p), name, usage, isFlag, isConfig)
    flag.DurationVar(p, name, value, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func DurationConfigVar(p *time.Duration, name string, value time.Duration, usage string) {
    isFlag := false
    isConfig := true
    configuration.Var(newDurationValue(value, p), name, usage, isFlag, isConfig)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func (c *ConfigoSet) Duration(name string, value time.Duration, usage string) *time.Duration {
    p := new(time.Duration)
    c.DurationVar(p, name, value, usage)
    return p
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func (c *ConfigoSet) DurationFlag(name string, value time.Duration, usage string) *time.Duration {
    p := new(time.Duration)
    c.DurationFlagVar(p, name, value, usage)
    return p
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func (c *ConfigoSet) DurationConfig(name string, value time.Duration, usage string) *time.Duration {
    p := new(time.Duration)
    c.DurationConfigVar(p, name, value, usage)
    return p
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func Duration(name string, value time.Duration, usage string) *time.Duration {
    return configuration.Duration(name, value, usage)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func DurationFlag(name string, value time.Duration, usage string) *time.Duration {
    return configuration.DurationFlag(name, value, usage)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func DurationConfig(name string, value time.Duration, usage string) *time.Duration {
    return configuration.DurationConfig(name, value, usage)
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func (c *ConfigoSet) Var(value flag.Value, name string, usage string, isFlag, isConfig bool) {
    // Remember the default value as a string; it won't change.
    config := &Configo{name, usage, value, value.String(), isFlag, isConfig}
    _, alreadythere := c.formal[name]
    if alreadythere {
        msg := fmt.Sprintf("%s flag redefined: %s", c.name, name)
        fmt.Fprintln(c.out(), msg)
        panic(msg) // Happens only if flags are declared with identical names
    }
    if c.formal == nil {
        c.formal = make(map[string]*Configo)
    }
    c.formal[name] = config
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value,
// which typically holds a user-defined implementation of Value. For instance,
// the caller could create a flag that turns a comma-separated string into a
// slice of strings by giving the slice the methods of Value; in particular,
// Set would decompose the comma-separated string into the slice.
// TODO This function does not appear to be used.
func Var(value flag.Value, name string, usage string, isFlag, isConfig bool) {
    configuration.Var(value, name, usage, isFlag, isConfig)

    if isFlag {
        flag.Var(value, name, usage)
    }
}

// NArg is the number of arguments remaining after flags have been processed.
func (c *ConfigoSet) NArg() int {
    return flag.NArg()
}

// NFlag returns the number of command-line flags that have been set.
func (c *ConfigoSet) NFlag() int {
    return flag.NFlag()
}

// Parse parses the command-line flags from os.Args[1:] and sets the values in
// this ConfigoSet.  Then the configuration file is parsed and any item that
// was not already set by the command line is set.  Must be called after all
// configuration options are defined and before conifguration options are
// accessed by the program.
func (c *ConfigoSet) Parse() (err error) {
    // Start by parsing the command line with the flag package and then set the
    // parsed values into the ConfigoSet.
    flag.Parse()
    flag.Visit(func(f *flag.Flag) {
        c.Set(f.Name, f.Value.String())
    })

    // Now parse the configuration file, but first create the config file if it
    // does not exist.  If that's the case we're all done and we can return.
    if _, err = os.Stat(c.path); err != nil {
        if !os.IsNotExist(err) {
            return
        }

        c.parsed = true
        err = c.WriteDefaultConfig(c.path)
        return
    }

    // Parse the config file, but only set the options that didn't appear on
    // the command line.
    if !c.parsed {
        var content []byte
        content, err = ioutil.ReadFile(c.path)
        if err != nil {
            return
        }

        for i, line := range strings.Split(string(content), "\n") {
            line = strings.TrimSpace(line)

            if len(line) > 0 && !strings.HasPrefix(line, "#") {
                var name, value string
                fields := strings.SplitN(line, c.delimiter, 2)
                if len(fields) != 2 {
                    errors.New(fmt.Sprintf("Invalid key%svalue pair in conifiguration file %s on line %d.\n", c.delimiter, c.path, i))
                }
                name = strings.TrimSpace(fields[0])
                value = strings.TrimSpace(fields[1])

                // Check if the item was already set from the command line.
                if _, exists := c.actual[name]; !exists {
                    // Is this even a valid config item?
                    config := c.Lookup(name)
                    if config == nil {
                        panic(errors.New("unknown configuration item"))
                    }

                    c.Set(name, value)
                }
            }
        }

        c.parsed = true
    }

    return
}

func Parse() error {
    return configuration.Parse()
}

/*
Parsed returns true if the configuration file and command-line flags have been
parsed.
*/
func (c *ConfigoSet) Parsed() bool {
    return c.parsed && flag.Parsed()
}

/*
PrintDefaults prints to standard error the default values of all defined
command-line flags.
*/
func (c *ConfigoSet) PrintDefaults() {
    c.VisitAll(func(config *Configo) {
        format := "  -%s=%s: %s\n"
        if _, ok := config.Value.(*stringValue); ok {
            // put quotes on the value
            format = "  -%s=%q: %s\n"
        }
        fmt.Fprintf(c.out(), format, config.Name, config.DefaultValue, config.Usage)
    })
}

// out returns the io.Writer where output should be sent.  Returns os.Stderr if
// no io.Writer has been set.
func (c *ConfigoSet) out() io.Writer {
    if c.output == nil {
        return os.Stderr
    }
    return c.output
}

/*
Set sets the value of the named configuration item.
*/
func (c *ConfigoSet) Set(name, value string) error {
    config, ok := c.formal[name]
    if !ok {
        return fmt.Errorf("no such configuration item %v", name)
    }
    err := config.Value.Set(value)
    if err != nil {
        return err
    }
    if c.actual == nil {
        c.actual = make(map[string]*Configo)
    }
    c.actual[name] = config
    return nil
}

/*
Visit visits the command-line flags in lexicographical order, calling fn for
each. It visits only those flags that have been set.
*/
func (c *ConfigoSet) Visit(fn func(*Configo)) {
    for _, config := range sortConfigs(c.actual) {
        fn(config)
    }
}

// Visit visits the command-line flags in lexicographical order, calling fn
// for each.  It visits only those flags that have been set.
func Visit(fn func(*Configo)) {
    configuration.Visit(fn)
}

/*
VisitAll visits the command-line flags in lexicographical order, calling fn for
each. It visits all flags, even those not set.
*/
func (c *ConfigoSet) VisitAll(fn func(*Configo)) {
    for _, config := range sortConfigs(c.formal) {
        fn(config)
    }
}

// VisitAll visits the configuration items in lexicographical order, calling fn
// for each.  It visits all flags, even those not set.
func VisitAll(fn func(*Configo)) {
    configuration.VisitAll(fn)
}

// sortConfigs returns the configuration items as a slice in lexicographical
// sorted order.
func sortConfigs(configs map[string]*Configo) []*Configo {
    list := make(sort.StringSlice, len(configs))
    i := 0
    for _, c := range configs {
        list[i] = c.Name
        i++
    }
    list.Sort()
    result := make([]*Configo, len(list))
    for i, name := range list {
        result[i] = configs[name]
    }
    return result
}

func SetDelimiter(d string) {
    configuration.delimiter = d
}

// Lookup returns the Configo structure of the named configo, returning nil if
// none exists.
func (c *ConfigoSet) Lookup(name string) *Configo {
    return c.formal[name]
}

// Lookup returns the Configo structure of the named configuration item,
// returning nil if none exists.
func Lookup(name string) *Configo {
    return configuration.formal[name]
}
