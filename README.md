# Extensible Flow Engine

## TODO

### Session

Importance of session, is to be able to serialize different sets of data with
different internal values, right now variables are being stored in session

## Features

* Develop new Operators
* Serialize graph of operations
* Distribute operations to serveral workers/servers
* HTTP Graphical UI Editor
* Serve http or grpc API
* Chain flow operators -- No needed since we can build operations,
  might be possible to build entires/methods from operators
* System containing more than one flow
* Define stages of the system

## Special operators

Idea: requesting the value to the flow instead of creating operators
allows to create a caching system to allow recursion with previous values

```go
f := flow.New(f);
v := f.Var([]float32{1,2,3}) // Init value
```

### New version

Simplified flow even more maintaining a func

### flowt2

Flow will create an array of operators with references as inputs
the builder will create the Graph

We should not allow user to directly pass Operation to the `flow.Op` method,
since the serialization will be hard then

define future inputs for operation groups

```go
g.Run([]O{op1,op2},1,2,3)
```

## Serialize

Grab all operators in a list
create a reference lookup table

## System

system combine several flows

Idea:

Develop and combine operators to create a function,

### Components

using ECS (entity component system) to extend each Operator, so if we have
UI we store UI information in UI component

### Describe components via code or serialization

Each node has components
create multiple channels to each node

CNN - combines convolution matrixes with images then sends to a regular
neural network

a Node could have input and properties?

## Backfetch:

perpare linkages as AddSource(node)
every time we call, it will call the results

```go
// Prototype
n := flow.AddNode(&node)
n.AddSource(&otherNode)
```
