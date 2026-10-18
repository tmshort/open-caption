# Translations of the guides

The guides in the parent folder (`usage-guide.md`, `venue-network-guide.md`, `https-certificate-guide.md`, `configuration-reference.md`) are written in English. The program shows
them in its Help in the six languages Setup supports; each translation lives in `docs/i18n/<code>/<guide>.md`
(`es`, `fr`, `de`, `it`, `pt`). The in-app help topics are translated the same way, in `web/help/<code>/`.

Rules for a translation (checked by the tests in `internal/help`):

- Same file names, and the first line is `# Title`.
- The same headings in the same order, the same fenced code blocks (copy them byte for byte: they are show file
  examples and commands), the same tables (same rows and columns), the same links. A translation lives two folders
  below the English guides, so a relative link needs more `../` there (`../../configuration-reference.md`,
  `../../../web/help/admin-getting-started.md`); a link to another guide that is translated is just its file name
  (`venue-network-guide.md`, the translated sibling). The tests compare the targets without the `../`, and check that
  every link in a translation reaches a file from where the translation is, so the links also work on GitHub.
- Links to a heading in the same document (`[text](#deployment-recommendations)`) must use the id of the *translated*
  heading: the id is the heading in lower case without accents, with other characters replaced by `-`.
- Do not translate commands, setting names, file names or keys typed in the program (`[loop]`, `font`, `config.yaml`,
  `Ctrl+C`); do translate the words around them. Button and menu names are quoted as they appear in that language's
  Setup.
