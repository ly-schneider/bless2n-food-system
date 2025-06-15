# Rules for the **LLM Code Assistant**

*(GitHub Copilot / Chat‑GPT‑style assistant inside VS Code for the **Rentro** repo – now based on the **Ignite** template by Infinite Red)*

---

## 1 — General Scope

1. **Single‑source truth** → Adhere to the *Master System Prompt for Rentro App* **and** these rules.
2. **Project only** → Generate content **only** for the Rentro mono‑repo (Expo x React Native). Ignore unrelated queries.
3. **Brand fidelity** → Enforce Rentro branding (colours, fonts, icon style, tone). Introduce **no** unapproved palettes or typography.

---

## 2 — Code‑Generation Conventions

| Topic                 | Rule                                                                                                                                             |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Language**          | **TypeScript** (`.tsx` / `.ts`) only.                                                                                                            |
| **Components**        | Functional components + Hooks; *never* use class components.                                                                                     |
| **Styling**           | Primary: **Tailwind CSS** via `nativewind` (Ignite‑compatible). Secondary: `styled-components` if Tailwind falls short. No inline style objects. |
| **Navigation**        | React Navigation v6; root navigator lives in `app/navigation/RootNavigator.tsx`. Route names come from each game module’s `routes.ts`.           |
| **State**             | Global game metadata with **MobX State Tree** (`app/models`) – default in Ignite. Local UI state with `useState`.                                |
| **File names**        | `PascalCase` components (`GameCard.tsx`); `kebab-case` assets (`logo-Rentro.svg`).                                                               |
| **Constants / Theme** | Define colours, spacing & fonts in `app/theme/colors.ts` & `app/theme/typography.ts`; **never** hard‑code.                                       |
| **Animations**        | Use `react-native-reanimated` v3; wrap gestures with `Reanimated.View`.                                                                          |
| **Assets**            | Place under `assets/<feature>/<name>.<ext>`; reference via Expo `require`.                                                                       |
| **Logging**           | Use `console.tron` (Ignite’s Reactotron wrapper) in development; disable in production (`app/config/reactotron.ts`).                             |

---

## 3 — Architecture & Folder Structure (Ignite‑style)

```
.
┣ app/
┃ ┣ components/          # Reusable UI primitives (buttons, cards…)
┃ ┣ screens/             # Generic app‑level screens (Home, Settings…)
┃ ┣ navigation/          # Root + dynamic game navigators
┃ ┣ models/              # MobX State Tree stores (gameStore, playerStore…)
┃ ┣ services/            # API / util services (if added later)
┃ ┣ theme/               # Rentro branding tokens, dark‑mode config
┃ ┣ utils/               # Generic helpers, formatters
┃ ┣ i18n/                # Translations (de.ts e.g.)
┃ ┗ app.tsx              # Ignite entry (wrapped by Expo)
┣ assets/                # Static assets (images, icons, fonts)
┣ plugins/               # Ignite plugins
┗ types/                 # TypeScript type definitions
```

*When the assistant creates a new file, it **must** conform to this tree.*

---

## 4 — UI / UX Rules

1. **White-mode only** – base background `#FFFFFF`.
2. **Tap targets ≥ 44 px** – verify with `react-native-accessibility-engine`.
3. **Micro‑animations** → subtle animations between page transitions and button presses.
4. **Iconography** → 2 px rounded‑stroke line icons via `lucide-react-native`.

---

## 5 — Testing & CI

* **Never** add tests when generating or refactoring logic.

---

## 6 — Assistant Behaviour

1. **Self‑check** → After generating code, re‑parse for TS errors & branding violations before returning.
2. **Ask when unsure** → Prompt the developer instead of guessing architectural decisions.
3. **No secrets** → Never output API keys, tokens, or personal data.
4. **Respect copyright** → Don’t paste licensed code verbatim; prefer original implementations.
5. **Explain briefly** → Prepend generated snippets with a concise comment (< 5 lines).

---

## 7 — Refactor & Documentation Rules

* When touching shared components, update **all** imports with VS Code organise‑imports.
* Keep JSDoc on every exported function/component.