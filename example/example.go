package main

import (
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/quincy/configo"
)

// Example 1: A single string flag called "species" with default value "gopher".
var species = configo.String("species", "gopher", "the species we are studying")

// Example 2: Two flags sharing a variable, so we can have a shorthand.
// The order of initialization is undefined, so make sure both use the
// same default value. They must be set up with an init function.
var gopherType string

func init() {
    const (
        defaultGopher = "pocket"
        usage         = "the variety of gopher"
    )
    configo.StringVar(&gopherType, "gopher_type", defaultGopher, usage)
    // shorthand version is not valid in the config file
    configo.StringVar(&gopherType, "g", defaultGopher, usage+" (shorthand)")
}

// Example 3: A user-defined flag type, a slice of durations.
type interval []time.Duration

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (i *interval) String() string {
    return fmt.Sprint(*i)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (i *interval) Set(value string) error {
    // If we wanted to allow the flag to be set multiple times,
    // accumulating values, we would delete this if statement.
    // That would permit usages such as
    //	-deltaT 10s -deltaT 15s
    // and other combinations.
    if len(*i) > 0 {
        return errors.New("interval flag already set")
    }
    for _, dt := range strings.Split(value, ",") {
        duration, err := time.ParseDuration(dt)
        if err != nil {
            return err
        }
        *i = append(*i, duration)
    }
    return nil
}

// Define a flag to accumulate durations. Because it has a special type,
// we need to use the Var function and therefore create the flag during
// init.

var intervalFlag interval

func init() {
    // Tie the command-line flag to the intervalFlag variable and
    // set a usage message.
    configo.Var(&intervalFlag, "deltaT", "comma-separated list of intervals to use between events", true, true)
}

// Example 4: some flag only options
var aliveFlag bool
var furryConfig bool

func init() {
    configo.BoolFlagVar(&aliveFlag, "alive", true, "set false to kill")
    configo.BoolConfigVar(&furryConfig, "furry", true, "furry or not")
}

func main() {
    if err := configo.Parse(); err != nil {
        panic(err)
    }

    fmt.Printf("species      = %s\n", *species)
    fmt.Printf("gopherType   = %s\n", gopherType)
    fmt.Printf("intervalFlag = %s\n", intervalFlag)
    fmt.Printf("alive        = %v\n", aliveFlag)
    fmt.Printf("furry        = %v\n", furryConfig)
}
