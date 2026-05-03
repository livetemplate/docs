---
title: "When to Reach for lvt-* (Decision Tree)"
---

# When to Reach for `lvt-*` (Decision Tree)

LiveTemplate has two tiers. **Tier 1** is plain HTML — `<form method="POST">`, `<a href="...">`, no JavaScript. **Tier 2** is the `lvt-*` attribute layer — debounced inputs, animations, optimistic UI, scoped DOM updates.

The framework's prevailing rule: **don't reach for Tier 2 unless Tier 1 has demonstrably hit its ceiling**. This page renders that rule as a decision tree you can walk before writing the next interaction.

```mermaid
flowchart TD
    Start([New interaction to implement]) --> Q1{Does the user submit<br/>a form or click a link?}

    Q1 -->|Yes| T1A[Use Tier 1: standard form/anchor.<br/>Action method on controller dispatches.]
    Q1 -->|No, it fires on input/keypress/focus| Q2

    Q2{Is debounce or<br/>throttle required?} -->|Yes| T2A
    Q2 -->|No| Q3

    Q3{Is optimistic UI<br/>required?<br/>e.g., grey out a button<br/>before server responds} -->|Yes| T2B
    Q3 -->|No| Q4

    Q4{Does the page need to<br/>re-render only ONE region<br/>without a full POST round-trip?} -->|Yes| T2C
    Q4 -->|No| Q5

    Q5{Is there a visual<br/>transition the user<br/>must see?<br/>e.g., row fade-in} -->|Yes| T2D
    Q5 -->|No| T1B

    T1A --> Done1[Done — pure server-side rendering]
    T1B[Use Tier 1 anyway.<br/>If a need surfaces later, layer on lvt-* THEN.] --> Done2[Done — keep it simple]

    T2A[lvt-on:input + Change handler] --> Done3
    T2B[lvt-el:setAttr or lvt-fx:* on the action source] --> Done3
    T2C[lvt-target / lvt-region wrapping the partial] --> Done3
    T2D[lvt-fx:animate / lvt-fx:highlight] --> Done3

    Done3[Done — layer on Tier 2 here, NOT site-wide]

    style T1A fill:#d4edda,stroke:#28a745
    style T1B fill:#d4edda,stroke:#28a745
    style T2A fill:#fff3cd,stroke:#ffc107
    style T2B fill:#fff3cd,stroke:#ffc107
    style T2C fill:#fff3cd,stroke:#ffc107
    style T2D fill:#fff3cd,stroke:#ffc107
```

## Why this matters

Every Tier 2 attribute is a small JavaScript dependency in the user's browser — a thing that can fail, can race, can accumulate. The framework was designed so the **escape hatch is opt-in per element**, not a global mode. A page can be 95% Tier 1 with one `<input lvt-on:input>` for live search and that's a feature, not a smell.

The most common Tier 2 over-application is **using `lvt-on:click` for a button that's inside a `<form>`**. If the form already has `name="Save"`, you don't need anything — submit dispatches. The Tier 2 layer is for cases where standard form/anchor semantics are inadequate, not for "everything I'd write in React."

## How this page works

This decision tree is a `mermaid` flowchart block — `\`\`\`mermaid` fence with a flowchart definition. Tinkerdown's bundled mermaid runtime renders it client-side after `DOMContentLoaded`. No template binding, no server round-trip — just a markdown block that becomes an SVG diagram.

For more on Tier 1 vs Tier 2 see the canonical reference at [Progressive Complexity](/guides/progressive-complexity).
