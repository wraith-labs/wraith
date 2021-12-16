# Wraith Architecture (WIP)
This document outlines how Wraith works in theory, how it communicates with C2, as well as how the codebase is structured.

## Index
- [Overview](#overview) - High-level outline of Wraith's design and design considerations
- [Features](#features) - Overview of Wraith's features and how they are implemented
- [Protocol](#protocol) - Detailed description of how Wraith communicates
- [Codebase Layout](#codebase-layout) - The layout of this repository

## Overview
Wraith is designed with flexibility, resilience and versatility in mind. This means that it should never need manual updating (that is, updating via original infection vectors) once deployed and should be able to deal with C2 outages, or the server being taken down altogether. Furthermore, Wraith is also designed to be modular, to allow for effortless expansion of its functionality, without familiarity with the entire codebase.

Wraith accomodates these requirements by utilising a unique architecture. The core component, libwraith, is tiny and lacks external dependencies, platform-specific code or cgo. It is designed to work as a library, meaning that it can be included in legitimate Go codebases with a minimal footprint, to provide a backdoor.

This architecture also allows for all basic and advanced functionality to be implemented as modules. Those can be included or excluded individually depending on requirements for each individual build. Wraith therefore adapts to be as covert or a feature-complete as desired.

As the C2 protocol is implemented as a plugin, it can be effortlessly switched out, or multiple protocols can work alongside eachother for maximum resilience. In practice, this means that command and control can take place over any protocol, including DNS which is extremely difficult to block.

## Features
- Library-like core:
  - Tiny
  - No external dependencies, platform-specific code or cgo
  - Can be included in legitimate Go codebases
- Extremely versatile:
  - Modules allow for adapting to each situation individually
  - Modules can be loaded and unloaded remotely as long as at least one pre-included module supports this functionality
- Difficult to detect:
  - Wraith is tailor-made to your requirements depending on which modules you include which makes it difficult to detect for antivirus software
- Core modules available, or write your own:
  - This repository contains a number of core modules within the stdmod package, which could be useful for general purpose usage
  - Modules are easy to write for more custom applications

## Protocol
Wraith is not tied to a specific protocol as this is dependent on modules. Wraith's communication works as follows:

- Internally, Wraith facilitates communication between modules by means of a SharedMemory instance. This is a thread-safe map-like structure which allows modules to write to "cells", read from them, and watch them for changes. The last of those also effectively makes SharedMemory a pub/sub queue. Between those features, SharedMemory provides a simple yet flexible way for modules to communicate.
- Some cells in SharedMemory are standardised and used for specific purposes. These include `SHM_TX_QUEUE` and `SHM_RX_QUEUE` which are used to queue messages received by Wraith or to be sent by Wraith.
- Special modules exist which have standardised names and are expected to carry out specific tasks. One of those is `MOD_COMMS_MANAGER`, which is responsible for managing the aforementioned queues. This module effectively governs the communication format as it is responible for reading from the queues, encoding/decoding and encrypting/decrypting the data. Note that it is only the format of messages which is decided by this module, not the protocol they are sent over.
- The comms manager module is then expected to pass the data on to other modules, depending on the queue the message was fetched from. Data from the TX queue, having been encoded and encrypted, should end up with a module capable of sending the data to C2. Meanwhile, messages read from the RX queue should end up with a module capable of processing them.
- It is then up to the modules to send off the data over their chosen protocol or to process it as they see fit.

Overall, the Wraith protocol is governed by the modules which are in use and entirely flexible. Modules even have the flexibility to bypass the comms manager altogether and make their own communication routes, though this is generally discouraged.

All that said, the default comms manager implemented as part of `stdmod` uses an encrypted JWT-based protocol.

## Codebase Layout
Due to the Wraith's modular architecture, the codebase is split into 2 main parts:
- libwraith:
  - The core which doesn't provide any functionality on its own. It can be included in other programs as a library as long as some modules are also included for it to execute.
- stdmod:
  - The standard module library - a list of modules maintained by the authors of Wraith which can be included directly from this repo. These are what provide actual functionality to Wraith and can serve as examples for writing your own custom modules.

In terms of the directory structure, that looks as follows:

- root (metadata and other non-code files)
  - docs (documentation and guides)
  - wraith (code)
    - libwraith (a flat filestructure containing all the necessary code for Wraith's core, split into multiple files)
    - stdmod (a collection of standard modules for Wraith)
    - vendor (external dependencies of stdmod included in the repo)