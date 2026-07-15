import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { extname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../content/docs/', import.meta.url));
const expectedPages = ['index.mdx', 'test.mdx'];

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
  assert.doesNotMatch(content, /\]\([^)]*\.md(?:#[^)]*)?\)/, `stale .md link in ${path}`);
}

console.log(`content check passed (${expectedPages.length} pages)`);
