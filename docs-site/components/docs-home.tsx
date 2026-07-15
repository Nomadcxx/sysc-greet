import { basePath } from '@/lib/shared';
import { ArrowRight } from 'lucide-react';
import Link from 'next/link';

const callsToAction = [
  ['Installation', 'Install sysc-greet for your compositor.', '/docs/getting-started/installation'],
  ['Quick Start', 'Configure greetd and launch the greeter.', '/docs/getting-started/quick-start'],
  ['Troubleshooting', 'Resolve common setup and runtime issues.', '/docs/getting-started/troubleshooting'],
] as const;

const footerLinks = [
  ['Install', '/docs/getting-started/installation'],
  ['Configure', '/docs/configuration/themes'],
  ['Compositors', '/docs/compositors/niri'],
  ['Develop', '/docs/development/architecture'],
] as const;

export function DocsHome() {
  return (
    <section className="docs-home-hero" aria-labelledby="docs-home-title" data-docs-home>
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
        Written in Go with the Bubble Tea framework, with compositor-aware setup and a greeter UI
        built for the console.
      </p>
      <nav className="docs-home-actions" aria-label="Get started">
        {callsToAction.map(([title, description, href]) => (
          <Link key={href} href={href}>
            <span>
              <strong>{title}</strong>
              <small>{description}</small>
            </span>
            <ArrowRight aria-hidden="true" />
          </Link>
        ))}
      </nav>
    </section>
  );
}

export function DocsHintStrip() {
  return (
    <nav className="docs-hint-strip" aria-label="Documentation sections">
      {footerLinks.map(([title, href]) => (
        <Link key={href} href={href}>
          {title}
        </Link>
      ))}
    </nav>
  );
}
