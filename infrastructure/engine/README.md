# Infrastructure/engine package
This is where the magic lies. There are currently 2 files, StubAuthzEngine and SpiceDbAuthzEngine. 

SpiceDbAuthzEngine.go is not implemented, but would be the technically highly coupled impl of using the spiceDB go client to make calls to spiceDB. I focused here on abstracting away the server, that's the only reason. To be implemented.