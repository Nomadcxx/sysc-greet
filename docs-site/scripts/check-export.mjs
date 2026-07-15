import assert from 'node:assert/strict';
import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const output = new URL('../out/', import.meta.url);
const basePath = process.env.GITHUB_ACTIONS === 'true' ? '/sysc-greet' : '';
const rootHtml = readFileSync(new URL('index.html', output), 'utf8');
const docsHtml = readFileSync(new URL('docs/index.html', output), 'utf8');
const articleHtml = readFileSync(
  new URL('docs/getting-started/installation/index.html', output),
  'utf8',
);
const searchPath = new URL('api/search', output);
const wordmarkPath = new URL('sysc-greet-wordmark.png', output);
const chunkRoot = fileURLToPath(new URL('_next/static/chunks/', output));
const sourceCss = readFileSync(new URL('../app/global.css', import.meta.url), 'utf8').replace(
  /\/\*[\s\S]*?\*\//g,
  '',
);
const css = files(fileURLToPath(new URL('_next/static/css/', output)))
  .map((path) => readFileSync(path, 'utf8'))
  .join('\n');

function files(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    return entry.isDirectory() ? files(path) : [path];
  });
}

assert.doesNotMatch(docsHtml, /Toggle Theme|light theme/i, 'dark-only export contains a theme toggle');
assert.match(css, /#ef233c/i, 'exported theme is missing the RAMA primary red');
assert.match(css, /#ff6678/i, 'exported theme is missing the accessible red text accent');
assert.match(css, /IBM Plex Sans Variable/i, 'exported theme is missing IBM Plex Sans');
assert.match(css, /Fira Code/i, 'exported theme is missing Fira Code');
assert.match(docsHtml, /logo\.png/, 'docs home is missing the existing sysc-greet logo');
assert.ok(
  docsHtml.includes(`src="${basePath}/logo.png"`),
  'docs logo is missing the deployment base path',
);
assert.ok(existsSync(wordmarkPath), 'exported sysc-greet wordmark asset is missing');
assert.ok(statSync(wordmarkPath).size > 0, 'exported sysc-greet wordmark asset is empty');
assert.match(
  docsHtml,
  /SEE YOU IN SPACE COWBOY/,
  'docs home is missing the approved tagline',
);
for (const [marker, label, route] of [
  ['install', 'INSTALL', '/docs/getting-started/installation/'],
  ['quick-start', 'QUICK START', '/docs/getting-started/quick-start/'],
  ['troubleshoot', 'TROUBLESHOOT', '/docs/getting-started/troubleshooting/'],
]) {
  const commandLink = new RegExp(
    `<a\\b(?=[^>]*\\bdata-home-command="${marker}")(?=[^>]*\\bhref="${basePath}${route}")[^>]*>\\s*/ ${label} /\\s*</a>`,
  );
  assert.ok(
    commandLink.test(docsHtml),
    `docs home is missing the ${label} command link`,
  );
}
assert.ok(
  docsHtml.includes(`src="${basePath}/sysc-greet-wordmark.png"`),
  'docs wordmark is missing the deployment base path',
);
assert.doesNotMatch(docsHtml, /docs-home-frame/, 'docs home still contains the old greeter frame');
assert.doesNotMatch(docsHtml, /\bslash-label\b/, 'docs home still contains the Documentation divider');
assert.doesNotMatch(
  docsHtml,
  /sysc-greet preview|assets\/showcase\.gif/,
  'docs home still contains the showcase preview',
);
assert.match(
  sourceCss,
  /\.docs-home-body\s*>\s*:where\(h2,\s*h3,\s*p,\s*ul,\s*ol,\s*pre,\s*figure,\s*table,\s*blockquote,\s*div\)\s*\{[^}]*width:\s*min\([^;]*52rem\)/,
  'home content blocks do not share the 52rem reading measure',
);
assert.match(
  sourceCss,
  /\.docs-home-body\s*>\s*:where\(img,\s*video\),\s*\.docs-home-body\s*>\s*:not\(\.docs-home-hero\)\s+:where\(img,\s*video\)\s*\{[^}]*width:\s*auto[^}]*height:\s*auto[^}]*margin-inline:\s*auto[^}]*object-fit:\s*contain/,
  'home content media sizing does not explicitly exclude the home hero',
);
assert.match(
  sourceCss,
  /\.docs-home-body\s*>\s*:where\(img,\s*video\)\s*\{[^}]*max-width:\s*min\(calc\(100%\s*-\s*2rem\),\s*52rem\)/,
  'direct home media does not use the centered 52rem measure',
);
assert.match(
  sourceCss,
  /\.docs-home-body\s*>\s*:not\(\.docs-home-hero\)\s+:where\(img,\s*video\)\s*\{[^}]*max-width:\s*100%/,
  'nested home content media is not constrained to its reading block',
);
assert.doesNotMatch(
  sourceCss,
  /\.docs-home-body\s+:where\(img,\s*video\)/,
  'a broad home media selector can override the masthead wordmark',
);
assert.match(
  sourceCss,
  /\.prose\s*>\s*h2\s*>\s*a\[data-card\]::before,\s*\.prose\s*>\s*h2\s*>\s*a\[data-card\]::after\s*\{[^}]*content:\s*['"]\/{10}['"]\s*\/\s*(?:''|"")[^}]*color:\s*inherit[^}]*font-size:\s*[\d.]+em/,
  'H2 title links are missing accessible relative ten-slash framing',
);
assert.match(
  sourceCss,
  /\.prose\s*>\s*h3\s*>\s*a\[data-card\]::before,\s*\.prose\s*>\s*h3\s*>\s*a\[data-card\]::after\s*\{[^}]*content:\s*['"]\/{4}['"]\s*\/\s*(?:''|"")[^}]*color:\s*inherit[^}]*font-size:\s*[\d.]+em/,
  'H3 title links are missing accessible relative four-slash framing',
);
assert.doesNotMatch(
  sourceCss,
  /(?:\.docs-home-body|\.prose:not\(\.docs-home-body\))\s*>\s*h[23]::(?:before|after)/,
  'slash framing is still attached outside the title link',
);
assert.match(
  sourceCss,
  /#nd-sidebar\s+\[data-search-full\]\s*\{[^}]*display:\s*none/,
  'sidebar duplicate full search is not hidden',
);
const hintRule = sourceCss.match(/\.docs-hint-strip\s*\{([^}]*)\}/)?.[1] ?? '';
assert.match(hintRule, /border-top:\s*1px\s+solid/, 'docs hint strip is missing its rail');
assert.doesNotMatch(hintRule, /(?:^|;)\s*border\s*:/, 'docs hint strip still has an enclosing border');
for (const route of [
  '/docs/getting-started/installation/',
  '/docs/getting-started/quick-start/',
  '/docs/getting-started/troubleshooting/',
]) {
  assert.ok(docsHtml.includes(`href="${basePath}${route}"`), `docs home is missing ${route}`);
}
for (const hint of ['Install', 'Themes', 'Compositors', 'Develop']) {
  assert.match(docsHtml, new RegExp(`>${hint}<`), `docs home is missing the ${hint} status hint`);
}
assert.match(
  docsHtml,
  /https:\/\/github\.com\/Nomadcxx\/sysc-greet/,
  'exported docs are missing the repository link',
);
assert.match(
  articleHtml,
  /blob\/development\/docs-site\/content\/docs\/getting-started\/installation\.md/,
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
assert.ok(
  rootHtml.includes(`url=${basePath}/docs/`),
  'site root does not redirect to documentation',
);

console.log('export check passed');
