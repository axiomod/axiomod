
# 📜 Decision Records (ADR) for Axiomod Framework

## Purpose

This folder contains **Architectural Decision Records (ADRs)** for the Axiomod Framework.

An ADR captures a **single, significant technical or architectural decision**, along with its context and consequences.  
Maintaining ADRs helps document why decisions were made, not just what was done.

## Folder Structure

- `ADR-000-template.md` — Use this template when writing new ADRs.
- `ADR-001-adopt-updated-project-structure.md` — The first ADR defining major structural changes.

## How to Create a New ADR

1. Copy `ADR-000-template.md` to a new file:

   ```plaintext
   cp ADR-000-template.md ADR-00X-title-of-decision.md
   ```

2. Increment the ADR number (`00X`).
3. Update the title, context, decision, and consequences.
4. Commit the new ADR alongside any code or documentation changes.

## Status Values

- **Proposed** — Still under review or discussion.
- **Accepted** — Approved and implemented (or agreed to implement).
- **Superseded** — Replaced by a newer ADR.
- **Deprecated** — No longer relevant or in use.

## Why Use ADRs?

- Ensure transparency for future contributors.
- Understand the reasoning behind critical architecture and technology choices.
- Track architectural evolution over time.

---

✅ **Start by reviewing ADR-001 for examples of how to format a new ADR.**
