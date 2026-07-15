import type { ComponentProps } from 'react';

const basePath = process.env.NEXT_PUBLIC_BASE_PATH ?? '';

export function Brand({ className, ...props }: ComponentProps<'span'>) {
  return (
    <span className={`sysc-brand ${className ?? ''}`} {...props}>
      <img src={`${basePath}/logo.png`} alt="" width="873" height="140" />
    </span>
  );
}
