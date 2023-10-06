# Package Descriptions

## cmd

Contains main file and command line flag parsing

## internal

Contains generic helper structs and functions: 

- broadcast: A utility packages for broadcasting channels
- infrastructure: Definition and handling of REST endpoint handlers

## pkg

Contains the core functionality and interfaces of this adapter:

- control: The core logic regarding essential resource types, e.g., pods and nodes
- interfaces: The REST interfaces for communication with MiSim and Kubernetes components
- misim: Misim specific data types and logic
- storage: Interfaces and structs for storing data in the adapter
