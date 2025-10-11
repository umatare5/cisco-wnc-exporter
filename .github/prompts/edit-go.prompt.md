---
description: "Edit Go code related to this project, ensuring consistency across files and adherence to specified rules."
mode: "agent"
model: "Claude Sonnet 4.5 (Preview)"
---

## Inputs

- Context: `${input:Context}`

## Baseline Instructions

- [copilot-instructions.md](../copilot-instructions.md)
- [go.instructions.md](../instructions/go.instructions.md)
- [go-prom-umatare5.instructions.md](../instructions/go-prom-umatare5.instructions.md)

## Environment

- [README.md](../../README.md)
- [about-labo-environment.md](./appendix/about-labo-environment.md)

## Instructions

${input:Context}
