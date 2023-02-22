# Domain package
Here we work only with Domain objects and interfaces, no technical impl details allowed.

Currently there are no domain objects, might merge with the other draft.

There is one domain service, a hello world example based on the hello world endpoint we had. in "handler", which for sure is bad naming and debatable.

Having this strictly technically-decoupled layer allows for strict separation of "technical concerns" and "business concerns", and as a result for not being tightly coupled to one specific technology, which usually pays out mid- to long term. 