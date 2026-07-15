import { basePath } from '@/lib/shared';
import Link from 'next/link';

const callsToAction = [
  ['install', '/ INSTALL /', '/docs/getting-started/installation'],
  ['quick-start', '/ QUICK START /', '/docs/getting-started/quick-start'],
  ['troubleshoot', '/ TROUBLESHOOT /', '/docs/getting-started/troubleshooting'],
] as const;

const tickerItems = [
  ['HYPRLAND', 'Your config worked yesterday. Fuck you, update available.'],
  ['HYPRLAND', 'Git bisect is your desktop environment.'],
  ['HYPRLAND', 'Your desktop has more animations than working protocols.'],
  ['HYPRLAND', 'You update once and spend Friday excavating your dotfiles.'],
  ['HYPRLAND', 'Maintainers ship the screenshots. Users debug the release.'],
  ['HYPRLAND', 'Maintainers closed your bug. The rewrite needs a different bug report.'],
  ['NIRI', 'Rust made the compositor memory-safe. Your wrists are on their own.'],
  ['NIRI', 'Infinite workspace, finite chance of finding the window you lost.'],
  ['NIRI', 'Horizontal scrolling turned into a fucking worldview.'],
  ['NIRI', 'Scrollable tiling makes fixed workspaces feel broken.'],
  ['NIRI', 'The rare new compositor whose big idea survives a full workday.'],
  ['NIRI', 'Wayland found a good reason to scroll sideways.'],
  ['GNOME', 'Developers removed the feature and published a manifesto about your mistake.'],
  ['GNOME', 'The settings panel has fewer options than the developers have opinions.'],
  ['GNOME', 'Your workflow failed ideological review.'],
  ['GNOME', 'Developers removed your button and called the empty space intentional.'],
  ['GNOME', 'Extensions let volunteers restore what maintainers removed on principle.'],
  ['GNOME', "Your right-click menu died for somebody's design system."],
] as const;

const tickerLabel = tickerItems.map(([target, roast]) => `[${target}] ${roast}`).join(' │ ');

const hintLinks = [
  ['Install', '/docs/getting-started/installation'],
  ['Themes', '/docs/configuration/themes'],
  ['Compositors', '/docs/compositors/niri'],
  ['Develop', '/docs/development/architecture'],
] as const;

export function DocsHome() {
  return (
    <section className="docs-home-hero" aria-label="sysc-greet documentation" data-docs-home>
      <div className="docs-home-masthead">
        <p className="docs-tagline">////////// SEE YOU IN SPACE COWBOY //////////</p>
        <div className="docs-home-wordmark-crop">
          <img
            className="docs-home-wordmark"
            src={`${basePath}/sysc-greet-wordmark.png`}
            alt="sysc-greet"
            width="2048"
            height="683"
          />
        </div>
      </div>
      <div className="docs-home-ticker" data-home-ticker role="note" aria-label={tickerLabel}>
        <div className="docs-home-ticker-track" aria-hidden="true">
          {[0, 1].map((copyIndex) => (
            <span className="docs-home-ticker-copy" key={copyIndex}>
              {tickerItems.map(([target, roast]) => (
                <span className="docs-home-ticker-item" key={`${target}-${roast}`}>
                  <span className="docs-home-ticker-target">[{target}]</span>
                  <span>{roast}</span>
                  <span className="docs-home-ticker-separator">│</span>
                </span>
              ))}
            </span>
          ))}
        </div>
      </div>
      <nav className="docs-home-actions" aria-label="Get started">
        {callsToAction.map(([marker, title, href]) => (
          <Link key={href} href={href} data-home-command={marker}>
            {title}
          </Link>
        ))}
      </nav>
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
