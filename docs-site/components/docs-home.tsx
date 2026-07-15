import { basePath } from '@/lib/shared';
import Link from 'next/link';

const callsToAction = [
  ['Install', '/docs/getting-started/installation'],
  ['Quick Start', '/docs/getting-started/quick-start'],
  ['Troubleshoot', '/docs/getting-started/troubleshooting'],
] as const;

const hintLinks = [
  ['Install', '/docs/getting-started/installation'],
  ['Themes', '/docs/configuration/themes'],
  ['Compositors', '/docs/compositors/niri'],
  ['Develop', '/docs/development/architecture'],
] as const;

export function DocsHome() {
  return (
    <section className="docs-home-hero" aria-labelledby="docs-home-title" data-docs-home>
      <div className="docs-home-frame">
        <p className="docs-tagline">////////// SEE YOU IN SPACE COWBOY //////////</p>
        <img
          className="docs-home-logo"
          src={`${basePath}/logo.png`}
          alt="sysc-greet"
          width="873"
          height="140"
        />
        <h1 id="docs-home-title">Graphical console greeter for greetd</h1>
        <p className="docs-home-summary">
          Written in Go with the Bubble Tea framework — themes, ASCII sessions, backgrounds, and
          compositor-aware setup.
        </p>
        <nav className="docs-home-actions" aria-label="Get started">
          {callsToAction.map(([title, href]) => (
            <Link key={href} href={href}>
              {title}
            </Link>
          ))}
        </nav>
      </div>
    </section>
  );
}

export function DocsHintStrip() {
  return (
    <nav className="docs-hint-strip" aria-label="Documentation sections">
      <span className="docs-hint-prefix">F1</span>
      {hintLinks.map(([title, href], index) => (
        <span key={href} className="docs-hint-item">
          {index > 0 ? <span className="docs-hint-sep">|</span> : null}
          <Link href={href}>{title}</Link>
        </span>
      ))}
    </nav>
  );
}
