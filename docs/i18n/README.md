# Translations of the guides

The guides in the parent folder (`usage-guide.md`, `venue-network-guide.md`) are written in English. The program shows
them in its Help in the six languages Setup supports; each translation lives in `docs/i18n/<code>/<guide>.md`
(`es`, `fr`, `de`, `it`, `pt`). The in-app help topics are translated the same way, in `web/help/<code>/`.

Rules for a translation (checked by the tests in `internal/help`):

- Same file names, and the first line is `# Title`.
- The same headings in the same order, the same fenced code blocks (copy them byte for byte: they are show file
  examples and commands), the same tables (same rows and columns), the same links.
- Links to a heading in the same document (`[text](#deployment-recommendations)`) must use the id of the *translated*
  heading: the id is the heading in lower case without accents, with other characters replaced by `-`.
- Do not translate commands, setting names, file names or keys typed in the program (`[loop]`, `font`, `config.yaml`,
  `Ctrl+C`); do translate the words around them. Button and menu names are quoted as they appear in that language's
  Setup.
