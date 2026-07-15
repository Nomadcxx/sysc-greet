import assert from 'node:assert/strict';
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const output = new URL('../out/', import.meta.url);
const basePath = process.env.GITHUB_ACTIONS === 'true' ? '/sysc-greet' : '';
const rootHtml = readFileSync(new URL('index.html', output), 'utf8');
const docsHtml = readFileSync(new URL('docs/index.html', output), 'utf8');
const articleHtml = readFileSync(new URL('docs/test/index.html', output), 'utf8');
const searchPath = new URL('api/search', output);
const chunkRoot = fileURLToPath(new URL('_next/static/chunks/', output));

function files(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    return entry.isDirectory() ? files(path) : [path];
  });
}

assert.doesNotMatch(docsHtml, /Toggle Theme|light theme/i, 'dark-only export contains a theme toggle');
assert.match(
  docsHtml,
  /https:\/\/github\.com\/Nomadcxx\/sysc-greet/,
  'exported docs are missing the repository link',
);
assert.match(
  articleHtml,
  /blob\/development\/docs-site\/content\/docs\/test\.mdx/,
  'article source link does not target the docs-site content',
);
assert.ok(statSync(searchPath).size > 100, 'static search payload is empty');
assert.ok(
  files(chunkRoot).some(
    (path) => path.endsWith('.js') && readFileSync(path, 'utf8').includes(`${basePath}/api/search`),
  ),
  'search client is missing the deployment base path',
);
assert.ok(rootHtml.includes(`href="${basePath}/docs/"`), 'site root does not link to documentation');

console.log('export check passed');
