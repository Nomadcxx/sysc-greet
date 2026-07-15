import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { extname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../content/docs/', import.meta.url));
const expectedPages = [
  'index.mdx',
  'getting-started/installation.md',
  'getting-started/quick-start.md',
  'getting-started/troubleshooting.md',
  'features/backgrounds-effects.md',
  'features/ascii-art.md',
  'features/wallpapers.md',
  'features/screensaver.md',
  'configuration/themes.md',
  'configuration/backgrounds.md',
  'configuration/keyboard-layout.md',
  'compositors/niri.md',
  'compositors/cagebreak.md',
  'compositors/sway.md',
  'compositors/hyprland.mdx',
  'development/architecture.md',
  'development/building.md',
  'development/testing.md',
];

function contentFiles(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    return entry.isDirectory() ? contentFiles(path) : [path];
  });
}

for (const page of expectedPages) {
  assert.doesNotThrow(() => readFileSync(join(root, page)), `missing documentation page: ${page}`);
}

for (const path of contentFiles(root).filter((file) => ['.md', '.mdx'].includes(extname(file)))) {
  const content = readFileSync(path, 'utf8');
  const rel = relative(root, path);
  assert.doesNotMatch(content, /\]\([^)]*\.md(?:#[^)]*)?\)/, `stale .md link in ${rel}`);
  assert.doesNotMatch(content, /^!!!/m, `legacy admonition syntax in ${rel}`);
  assert.doesNotMatch(content, /docs-src|mkdocs/i, `stale documentation system reference in ${rel}`);
}

console.log(`content check passed (${expectedPages.length} pages)`);
