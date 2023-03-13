# application layer

boiled down, the "sticking it together" layer between api, domain, technical impls etc.

Also contains builders that build the variances (engine, server, ...) based on the given configuration. so this can be used inside the domain and elsewhere if needed, without tight coupling to the technical implementation. Currently, only "Serve" is used for demo purposes.

NOTE: this is where I am very unsure myself, if implemented right. I usually used it to bootstrap the application, set the actual impls I wanted based on an abstracted away config. Called it "bootstrap" often in the past.

Further reading: see the answer from Travis Parks here: https://softwareengineering.stackexchange.com/questions/140999/application-layer-vs-domain-layer