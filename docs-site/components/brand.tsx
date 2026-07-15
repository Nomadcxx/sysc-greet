import type { ComponentProps } from 'react';

const basePath = process.env.NEXT_PUBLIC_BASE_PATH ?? '';
const blockWordmark = ` ▄▄▄▄▄▄▄ ▄▄    ▄▄   ▄▄▄▄▄▄▄  ▄▄▄▄▄▄▄     ▄▄    ▄▄    ▄▄    ▄▄
██▀▀▀▀▀▀ ██▄  ▄██  ██▀▀▀▀▀▀ ██▀▀▀▀▀▀    ▄██   ▄██   ▄██   ▄██
▀██████▄  ▀████▀   ▀██████▄ ██        ▄██▀  ▄██▀  ▄██▀  ▄██▀
▄▄▄▄▄▄██    ██     ▄▄▄▄▄▄██ ██▄▄▄▄▄▄ ██▀   ██▀   ██▀   ██▀
▀▀▀▀▀▀▀     ▀▀     ▀▀▀▀▀▀▀   ▀▀▀▀▀▀▀ ▀▀    ▀▀    ▀▀    ▀▀`;

export function Brand({ className, ...props }: ComponentProps<'span'>) {
  return (
    <span className={`sysc-brand ${className ?? ''}`} {...props}>
      <span className="sysc-brand-ascii" aria-hidden="true">
        {blockWordmark}
      </span>
      <img
        className="sysc-brand-compact"
        src={`${basePath}/logo.png`}
        alt=""
        width="873"
        height="140"
      />
    </span>
  );
}
