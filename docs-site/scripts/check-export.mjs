import assert from 'node:assert/strict';
import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const output = new URL('../out/', import.meta.url);
const repositoryRoot = new URL('../../', import.meta.url);
const basePath = process.env.GITHUB_ACTIONS === 'true' ? '/sysc-greet' : '';
const rootHtml = readFileSync(new URL('index.html', output), 'utf8');
const docsHtml = readFileSync(new URL('docs/index.html', output), 'utf8');
const docsWorkflow = readFileSync(new URL('.github/workflows/docs.yml', repositoryRoot), 'utf8');
const readme = readFileSync(new URL('README.md', repositoryRoot), 'utf8');
const postinstall = readFileSync(new URL('scripts/postinstall.sh', repositoryRoot), 'utf8');
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
const docsLayout = readFileSync(new URL('../app/docs/layout.tsx', import.meta.url), 'utf8');
const sharedLayout = readFileSync(new URL('../lib/layout.shared.tsx', import.meta.url), 'utf8');
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
assert.match(docsHtml, /data-home-ticker/, 'docs home is missing the compositor ticker');
assert.doesNotMatch(
  docsHtml,
  /Graphical console greeter for greetd|Written in Go with the Bubble Tea framework/,
  'docs home still contains the discarded landing intro copy',
);
assert.doesNotMatch(docsHtml, /docs-home-intro/, 'docs home still wraps commands in the old intro');
assert.match(
  docsHtml,
  /data-home-ticker[\s\S]*?<\/div><nav class="docs-home-actions"/,
  'docs commands do not follow the ticker directly',
);
assert.doesNotMatch(
  docsHtml,
  /docs-home-transition/,
  'docs home still contains the plain masthead separator',
);
for (const roast of [
  '[HYPRLAND] Your config worked yesterday. Fuck you, update available.',
  '[HYPRLAND] Maintainers closed your bug. The rewrite needs a different bug report.',
  '[NIRI] Rust made the compositor memory-safe. Your wrists are on their own.',
  '[NIRI] Wayland found a good reason to scroll sideways.',
  '[GNOME] Developers removed the feature and published a manifesto about your mistake.',
  "[GNOME] Your right-click menu died for somebody's design system.",
]) {
  assert.ok(docsHtml.includes(roast), `docs ticker is missing: ${roast}`);
}
assert.match(
  sourceCss,
  /\.docs-home-wordmark-crop\s*\{[^}]*height:\s*clamp\([^;]*7\.25rem\)/,
  'docs wordmark crop does not expose the descenders',
);
assert.match(
  sourceCss,
  /\.docs-home-wordmark\s*\{[^}]*object-position:\s*center\s+55%/,
  'docs wordmark crop window is not lowered',
);
assert.match(
  sourceCss,
  /\.docs-home-actions\s+a\s*\{[^}]*border:\s*(?:0|none)[^}]*background:\s*none[^}]*color:\s*#edf2f4[^}]*text-decoration:\s*none/i,
  'docs home commands are not plain RAMA-white text links',
);
assert.match(
  sourceCss,
  /\.docs-home-actions\s+a:hover,\s*\.docs-home-actions\s+a:focus-visible\s*\{[^}]*color:\s*#ff6678/i,
  'docs home commands do not use RAMA red for interaction feedback',
);
assert.match(
  docsLayout,
  /sidebar=\{\{\s*defaultOpenLevel:\s*1\s*\}\}/,
  'docs layout does not open top-level sidebar categories by default',
);
assert.ok(
  sharedLayout.includes('className="sidebar-slash-mark"') &&
    sharedLayout.includes('MENU////////////'),
  'sidebar navigation title is missing the approved menu label',
);
assert.match(
  sourceCss,
  /\.sidebar-slash-mark\s*\{[^}]*color:\s*#ff6678[^}]*font-family:\s*['"]Fira Code['"]/i,
  'sidebar slash motif is missing its terminal treatment',
);
assert.match(
  sourceCss,
  /#nd-sidebar\s+a:has\(>\s*\.sidebar-slash-mark\)\s*\{[^}]*margin-inline-end:\s*0\.35rem/i,
  'sidebar slash motif is not grouped with the collapse control',
);
assert.match(
  sourceCss,
  /#nd-sidebar\s+\[data-(?:open|closed)\]\s*>\s*button\[aria-expanded\][\s\S]*?text-transform:\s*uppercase/i,
  'sidebar category triggers are missing their uppercase red treatment',
);
assert.match(
  sourceCss,
  /#nd-sidebar\s+\[data-open\]\s*>\s*\[data-open\]\s*\{[^}]*border-inline-start:\s*1px\s+solid/i,
  'sidebar child links are missing their guide rail',
);
assert.match(
  sourceCss,
  /#nd-sidebar\s+a\[data-active=['"]true['"]\]\s*\{[^}]*border-radius:\s*0[^}]*border-inline-start:\s*2px\s+solid\s+#ff6678[^}]*background:\s*rgb\(255\s+102\s+120/i,
  'sidebar active link is missing its square red edge treatment',
);
assert.match(
  sourceCss,
  /@keyframes\s+docs-home-ticker-scroll/,
  'docs ticker is missing its scrolling animation',
);
assert.match(
  sourceCss,
  /@media\s*\(prefers-reduced-motion:\s*reduce\)[\s\S]*?\.docs-home-ticker-track\s*\{[^}]*animation:\s*none/,
  'docs ticker is missing its reduced-motion fallback',
);
assert.match(docsHtml, /▄▄▄▄▄▄▄/, 'header is missing the supplied block wordmark');
assert.match(docsHtml, /▀██████▄/, 'header block wordmark is incomplete');
assert.match(
  docsHtml,
  /img\.shields\.io\/github\/stars\/Nomadcxx\/sysc-greet/,
  'header is missing the live GitHub stars badge',
);
assert.match(
  docsHtml,
  /img\.shields\.io\/badge\/Go-1\.25\.1-00ADD8/i,
  'header is missing the Go version badge',
);
assert.match(
  sourceCss,
  /\.greeter-header\s*\{[^}]*display:\s*grid[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s+minmax\(520px,\s*720px\)\s+minmax\(0,\s*1fr\)/,
  'desktop header does not centre the doubled search track',
);
assert.match(
  sourceCss,
  /\.greeter-search-full\s*>\s*button\s*\{[^}]*border-radius:\s*999px/,
  'desktop search control is not pill-shaped',
);
assert.match(
  sourceCss,
  /@media\s*\(max-width:\s*767px\)[\s\S]*?\.sysc-brand-ascii\s*\{[^}]*display:\s*none[\s\S]*?\.sysc-brand-compact\s*\{[^}]*display:\s*block/,
  'mobile header does not swap the block wordmark for the compact logo',
);
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
assert.doesNotMatch(docsWorkflow, /mkdocs|docs-src|setup-python|pip install/i, 'docs workflow still uses MkDocs');
assert.match(docsWorkflow, /docs-site\/\*\*/, 'docs workflow does not watch the Fumadocs source');
assert.match(docsWorkflow, /actions\/setup-node@v\d+/, 'docs workflow does not configure Node');
assert.match(docsWorkflow, /node-version:\s*['"]?22['"]?/, 'docs workflow does not use Node 22');
assert.match(docsWorkflow, /run:\s*npm ci/, 'docs workflow does not install locked dependencies');
assert.match(docsWorkflow, /run:\s*npm run check/, 'docs workflow does not run the full docs check');
assert.match(docsWorkflow, /working-directory:\s*docs-site/g, 'docs workflow commands do not run in docs-site');
assert.match(docsWorkflow, /path:\s*\.\/docs-site\/out/, 'docs workflow does not upload the static export');
assert.ok(!existsSync(new URL('mkdocs.yml', repositoryRoot)), 'legacy mkdocs.yml still exists');
assert.ok(!existsSync(new URL('docs-src/', repositoryRoot)), 'legacy docs-src still exists');
assert.doesNotMatch(
  readme,
  /nomadcxx\.github\.io\/sysc-greet\/getting-started\//i,
  'README still links to the pre-Fumadocs installation route',
);
assert.match(
  readme,
  /nomadcxx\.github\.io\/sysc-greet\/docs\/getting-started\/installation\//i,
  'README is missing the Fumadocs installation route',
);
assert.doesNotMatch(
  postinstall,
  /nomadcxx\.github\.io\/sysc-greet\/compositors\//i,
  'postinstall still links to the pre-Fumadocs compositor route',
);
assert.match(
  postinstall,
  /nomadcxx\.github\.io\/sysc-greet\/docs\/compositors\/cagebreak\//i,
  'postinstall is missing the Fumadocs Cagebreak route',
);

console.log('export check passed');
