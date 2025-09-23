---
description: "Additional documentation and content creation standards"
applyTo: "**/*.md"
---

# Additional Documentation and Content Creation Standards

GitHub Copilot **MUST** follow these instructions when generating or modifying Markdown files in this repository.

### Scope & Metadata

- **Last Updated**: 2025-08-10
- **Precedence**: 1. `copilot-instructions.md` (Global) → 2. `markdown.instructions.md` (Community) → 3. `markdown-umatare5.instructions.md` (This)
- **Compatibility**: [CommonMark](https://commonmark.org/) / [GitHub Flavored Markdown (GFM)](https://github.github.com/gfm/)
- **Style Base**: [Microsoft Writing Style Guide](https://learn.microsoft.com/style-guide/) and repository-specific conventions
- **Goal**: High-quality, maintainable, user-friendly Markdown documentation.

## Documentation Practices & Style

Apply the following style and best practices to all Markdown files **in addition to** the base Markdown Content Rules:

- MD-001 (**MUST**) **Readability:**
  Write in clear, concise language. Use short paragraphs and ensure sufficient whitespace between elements like sections, lists, and code blocks to improve readability.

- MD-002 (**SHOULD**) **Headings & Emojis:**
  Use a single H1 (`#`) for the main document title.
  Prefix H2 (`##`) and H3 (`###`) headings with a relevant emoji to visually represent the content.

- MD-003 (**MUST**) **Text Formatting:**
  Use `**bold**` text to emphasize key terms, warnings, or important notes.
  Use `` `inline code` `` for file paths, variable names, commands, and other code-related terms.

- MD-004 (**MUST**) **Consistent Structure:**
  Maintain a logical and consistent document structure across all files.
  A typical structure includes an introduction, setup/usage sections, detailed guides, and references.

- MD-005 (**MUST**) **Style & Linting:**
  Ensure all scripts pass `markdownlint-cli2`.

- MD-006 (**SHOULD**) **Use GitHub Alerts for Emphasis instead of standard blockquotes:**
  Use GitHub-flavored Markdown alerts (e.g., `> [!CAUTION]`, `> [!WARNING]`) to draw attention to critical information such as security risks or important usage notes if needed.

- MD-007 (**SHOULD**) **Collapsible Sections:**
  Use `<details>` and `<summary>` tags to enclose lengthy, non-critical content like sample outputs or verbose examples, improving overall document readability.
