# Wraith Architecture (WIP)
This document outlines how Wraith works in theory, and how it communicates with its C2, as well as how the codebase is structured.

## Index
- [Overview](#overview) - High-level outline of Wraith's design and design considerations
- [Features](#features) - Overview of Wraith's features and how they are implemented
- [Protocol](#protocol) - Detailed description of how Wraith communicates
- [Codebase Layout](#codebase-layout) - The layout of this repository

## Overview
Wraith is designed with flexibility, resilience and versatility in mind. This means that it should never need updating once deployed and should be able to deal with C2 outages, or the server being taken down altogether. Furthermore, Wraith is also designed to be modular, to allow for effortless expansion of its functionality, without familiarity with the entire codebase.

Wraith accomodates these requirements by utilising a unique architecture. The core component, libwraith, is tiny and lacks external dependencies, platform-specific
code or cgo. It is designed to work as a library, meaning that it can be included in legitimate Go codebases with a minimal footprint to provide a backdoor.

This architecture also allows for all basic and advanced functionality to be implemented as plugins. Those can be included or excluded individually depending on
requirements for each individual build. Wraith therefore adapts to be as covert or a feature-packed as desired.

As the C2 protocol is implemented as a plugin, it can be effortlessly switched out, or multiple protocols can work alongside eachother for maximum resilience.
In practice, this means that command and control can take place over any protocol, including DNS which is extremely difficult to block.

## Features
- Library-like core:
  - Tiny
  - No external dependencies, platform-specific code or cgo
  - Can be included in legitimate Go codebases
- Extremely versatile:
  - Plugins allow for adapting to each situation individually
  - Plugins can be loaded and unloaded remotely as long as at least one pre-built plugin supports this functionality
- Difficult to detect:
  - Wraith is tailor-made to your requirements depending on which plugins you include which makes it difficult to detect for antivirus software
- Core plugins available, or write your own:
  - This repository contains a number of core plugins which could be useful for general purposes
  - Plugins are easy to write for more custom applications

## Protocol
Wraith is not tied to a specific protocol as this is dependent on plugins. Wraith's communication is split into 4 components:
- Rx Plugins:
  - These are responsible for receiving data. This could be by polling a HTTP server, regularly checking a DNS record, listening on a socket,
  reading a specific part of the disk, or any other means of receiving data.
- Tx Plugins:
  - These are responsible for transmitting data. Once again, this could be achieved by means of a HTTP request, TCP socket connection, email or
  anything else.
- Proto Language Plugins:
  - These encode/decode data for/from Tx/Rx plugins. The Tx and Rx plugins are "dumb" in that they only understand their "carrier" protocol (HTTP, TCP, SMTP)
  but not the actual payload. Meanwhile, Language plugins are responsible for understanding the format of the payload and decoding it into a map which
  is then useable to Wraith. At this layer is where things like payload encryption, payload signing, payload signature verification etc., all take place.
- Proto Part Plugins:
  - These are perhaps the smartest of all plugins. They are dedicated to handling individual keys of the map produced by Language plugins. The value of that
  map is their argument. For instance, a command plugin may take the value of its key and execute it in a system shell; then send the result via a Tx plugin
  back to the source of the command. Each instance of that plugin also has a key-value store which it shares with other plugins per received payload. This allows
  Part plugins to interact. These plugins are also the most versatile because they can do anything with their arguments and can execute any code they like, including
  registering new plugins or re-installing Wraith.

## Codebase Layout
Due to the Wraith's modular architecture, the codebase is split into 2 parts:
- libwraith:
  - The core which doesn't provide any functionality on its own. It can be included in other programs as a library as long as some modules are also included for it
  to execute.
- stdmod:
  - The standard module library - a list of modules maintained by the authors of Wraith which can be included directly from this repo. These are what provide actual
  functionality to Wraith and can serve as examples for writing your own custom modules.

In terms of the file layout, that looks as follows:

- root (metadata and other non-code files)
  - wraith (the Wraith codebase)
    - libwraith (a flat filestructure containing all the necessary code for Wraith's core split into multiple files)
    - stdmod (a collection of standard modules for Wraith)
      - mod_lang (Proto Lang modules)
      - mod_part (Proto Part modules)
      - mod_rx (Rx modules)
      - mod_tx (Tx modules)