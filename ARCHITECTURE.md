# Wraith Architecture (WIP)
This document outlines how Wraith works in theory, and how it communicates with its C2, as well as how the codebase is structured.

## Index
- [Overview](#overview) - High-level outline of Wraith's design and design considerations
- [Functionality](#functionality) - Overview of Wraith's features and how they are implemented
- [Protocol](#protocol) - Detailed description of how Wraith communicates
- [Codebase Layout](#codebase-layout) - The layout of this repository

## Overview
Wraith is designed with flexibility and resilience in mind. This means that it should never need updating once deployed and should be able to deal with C2 outages, or the server being taken down altogether. Furthermore, Wraith is also designed to be modular, to allow for effortless expansion of its functionality, without familiarity with the entire codebase.

The Wraith C2 protocol accomodates these requirements, by building on top of JSON Web Tokens (JWT). This allows Wraiths to fetch commands from their peers, not just the C2 server, as each command is signed and can therefore be verified individually. The protocol can work on top of multiple others, which are commonly left unblocked across firewalls, such as HTTP(S) or DNS and, in theory, also over non-standard transmission methods like sound waves using a computer's speaker and microphone, in order to bridge the air gap.

## Functionality

## Protocol

## Codebase Layout
Due to the Wraith's modular architecture, each component of the Wraith is a separate submodule. These are then used in `wraith.go` which is the main module.

The submodules are laid out like so:
- **config**:
  - **`skeleton.go`** - *A structure defining the layout of the Wraith's config*
  - **`config.go`** - *The actual configuration file where options should be set before compiling*
- **constants**:
  - **`constants.go`** - *Various constants used by the Wraith, such as its current version*
- **comms**:
  - **channels**:
    - **`transmitters`** - *Modules (one per file) for sending data to C2*
    - **`receivers`** - *Modules (one per file) for receiving data and commands from C2*
  - **`comms.go`** - *The module responsible for managing transmitter and receiver modules*
- **`vendor`** - *Dependencies*
- **`main.go`** - *Main Wraith file*