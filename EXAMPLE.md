# Mermaid Complexity Detection Example

This file demonstrates how glow handles complex mermaid diagrams that exceed the ASCII renderer's capabilities.

## Simple Diagram (renders correctly)

```mermaid
graph LR
    A[Start] --> B{Decision}
    B -->|Yes| C[OK]
    B -->|No| D[Cancel]
```

## Complex Diagram (too complex for ASCII)

The diagram below has multiple subgraphs with cross-subgraph edges, which causes layout issues in the mermaid-ascii library (node duplication, garbled output).

```mermaid
flowchart LR
    subgraph SENSE[Sense]
        cam[Camera]
        unity[Unity]
        tests[Tests]
    end
    subgraph PROCESS[Process]
        perc[Perception]
        backend[Backend]
        commander[Commander]
    end
    subgraph CONTROL[Control]
        gimbal[Gimbal]
        servo[Servo]
    end
    subgraph OUTPUT[Output]
        frontend[Frontend]
        audio[Audio]
    end

    cam --> perc
    cam --> backend
    perc --> commander
    unity --> commander
    tests --> commander
    backend --> commander
    commander --> gimbal
    gimbal --> servo
    servo --> gimbal
    backend --> frontend
    commander --> audio
```

## Sequence Diagram (renders correctly)

```mermaid
sequenceDiagram
    Client->>Server: Request
    Server-->>Client: Response
```
