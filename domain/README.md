# Domain package
Here we work only with Domain objects and interfaces, no technical impl details allowed.

Having this strictly technically-decoupled layer allows for strict separation of "technical concerns" and "business concerns", and as a result for not being tightly coupled to one specific technology, which usually pays out mid- to long term. 

Package "contracts" contains interfaces to technical implementations of business capabilities.

Package "services" contains domain services that use the interfaces from "contract".

This package holds the models / domain entities and value objects.