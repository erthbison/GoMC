# Module Guide

Go-MC uses a modular design where each module has a clearly defined interface and where it is easy to swap between different implementations of a module.
This means that Go-MC can easily be used when designing new or testing scheduling algorithms.
It can also easily be customized to support with the framework you use in your implementation or the required abstraction. 

Go-MC uses four modules: the Schedule, the State Manager, The Failure Manager and the Checker. 
Each of the modules can be customized to provide the necessary behavior for the simulation.
In addition to these four modules Go-MC uses Event Managers which are inserted into the algorithm and are used to record and control the events.
Custom Event Managers that can be used with new primitives and frameworks can also be created and used. 

This document will provide an introduction to how to create own implementations of the modules and Event Managers. 

Content

- [Scheduler](#scheduler)
- [State Manager](#state-manager)
- [Failure Manager](#failure-manager)
- [Checker](#checker)
- [Event Manager](#event-manager)

## Scheduler

## State Manager

## Failure Manager

## Checker

## Event Manager
