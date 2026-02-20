# BFS Docs

Documentation site for the Bless2n Food System — built with Fumadocs on Next.js, covering admin handbooks, developer guides, and infrastructure documentation.

## Architecture

Built on [Fumadocs](https://fumadocs.dev), a documentation framework for Next.js that supports MDX content with React components, full-text search, and static generation.

### Content Sections

| Section | Path | Audience |
|---------|------|----------|
| Admin-Handbuch | `content/admin-handbuch/` | System administrators (German) |
| Developer Guide | `content/developer-guide/` | Developers |
| Infrastruktur | `content/infrastruktur/` | DevOps / Infrastructure (German) |

## Tech Stack

| Category | Technology |
|----------|-----------|
| Framework | Next.js 16 |
| Docs Engine | Fumadocs 16 (core + UI + MDX) |
| UI | React 19, Tailwind CSS 4 |
| Content | MDX with custom components |
| Search | Built-in API route |
| Icons | Lucide React |
| Package Manager | pnpm |

## Project Structure

```
bfs-docs/
├── content/
│   ├── admin-handbuch/       Admin user guide (German)
│   ├── developer-guide/      Developer documentation
│   └── infrastruktur/        Infrastructure docs (German)
├── app/
│   ├── page.tsx              Home page
│   └── api/search/route.ts   Search endpoint
├── lib/
│   ├── source.ts             Content source adapter
│   └── layout.shared.tsx     Shared layout options
├── .source/                  Generated source cache
├── source.config.ts          Fumadocs MDX configuration
├── mdx-components.tsx        MDX component overrides
├── next.config.mjs           Next.js configuration
├── Dockerfile                Production build
└── package.json
```

## Development

```bash
pnpm install
pnpm dev              # http://localhost:3000
```

## Commands

```bash
pnpm dev              # Development server
pnpm build            # Production build
pnpm lint             # ESLint
```

## Adding Content

1. Create an MDX file in the appropriate `content/` subdirectory
2. Add frontmatter (title, description)
3. The page is automatically picked up by Fumadocs and added to navigation

Customize frontmatter schema in `source.config.ts`. See the [Fumadocs MDX docs](https://fumadocs.dev/docs/mdx) for advanced features.
