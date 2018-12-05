package main

import (
	"flag"
	"github.com/racerxdl/radioserver/protocol"
)

var ServerVersion = protocol.Version{
	Major:    2,
	Minor:    0,
	Revision: 1700,
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
