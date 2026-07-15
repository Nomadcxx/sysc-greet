'use client';

import { Brand } from '@/components/brand';
import { SidebarTrigger } from 'fumadocs-ui/layouts/docs/slots/sidebar';
import { FullSearchTrigger, SearchTrigger } from 'fumadocs-ui/layouts/shared/slots/search-trigger';
import { GitFork, PanelLeft } from 'lucide-react';
import Link from 'next/link';
import type { ComponentProps } from 'react';

export function GreeterHeader({ className, ...props }: ComponentProps<'header'>) {
  return (
    <header {...props} className={`greeter-header ${className ?? ''}`}>
      <Link href="/docs" className="greeter-brand" aria-label="sysc-greet documentation home">
        <Brand />
      </Link>
      <div className="greeter-header-actions">
        <div className="greeter-search-full">
          <FullSearchTrigger hideIfDisabled />
        </div>
        <SearchTrigger hideIfDisabled aria-label="Open search" className="greeter-search-icon" />
        <a
          className="greeter-icon-link"
          href="https://github.com/Nomadcxx/sysc-greet"
          aria-label="sysc-greet on GitHub"
          title="GitHub repository"
        >
          <GitFork aria-hidden="true" />
        </a>
        <SidebarTrigger className="greeter-sidebar-trigger" aria-label="Open navigation">
          <PanelLeft aria-hidden="true" />
        </SidebarTrigger>
      </div>
    </header>
  );
}
