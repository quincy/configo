configo
=======

Simple configuration files _AND_ command line parsing for Go.

The same API provided by the flag package, extended for use with configuration files.  Use a single function to setup a command line flag and configuration file option simultaneously.

Get the source

    $ go get githib.com/quincy/configo

Usage
-----

    // file: example.go
    package main
    
    import "github.com/quincy/configo"
    
    var species = configo.String("species", "gopher", "the species we are studying", true, true)
    //                                                                                 ^     ^
    //                                This option can be set from the command line ----+     |
    //                                                                                       |
    //                                This option can be set from the configuration file ----+

    configo.SetDelimiter(":")  // Default is "="
    configo.Parse()
    
    fmt.Printf("Value of species is %s\n", species)


The first time your program is run a default config file will be created with
all options set to defaults, commented with the usage text.

    $ go run example.go
    Value of species is gopher
    $ cat ~/.examplerc
    # Default config file for example
    # Written on 07 Aug 13 20:15 -0600
    
    # the species we are studying
    species:gopher
    
The command line arguments override those found in the config file.

    $ go run example.go -species=mole
    Value of species is mole

See example/example.go for more complicated examples.
