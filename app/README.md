# application layer

see the answer from Travis Parks here: https://softwareengineering.stackexchange.com/questions/140999/application-layer-vs-domain-layer

boiled down, the translation layer between domain, technical impls etc.

NOTE: this is where I am very unconfident myself. I usually used it to bootstrap the application, set the actual impls I wanted based on an abstracted away config. Which is perfectly possible here, see app.go and think of a non-example config containing e.g. the technical implementation key for the server, the repository implementations, ports etc..

To be discussed later :)