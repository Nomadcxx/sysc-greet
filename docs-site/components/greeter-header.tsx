'use client';

import { Brand } from '@/components/brand';
import { SidebarTrigger } from 'fumadocs-ui/layouts/docs/slots/sidebar';
import { FullSearchTrigger, SearchTrigger } from 'fumadocs-ui/layouts/shared/slots/search-trigger';
import { PanelLeft } from 'lucide-react';
import Link from 'next/link';
import type { ComponentProps } from 'react';

export function GreeterHeader({ className, ...props }: ComponentProps<'header'>) {
  return (
    <header {...props} className={`greeter-header ${className ?? ''}`}>
      <Link href="/docs" className="greeter-brand" aria-label="sysc-greet documentation home">
        <Brand />
      </Link>
      <div className="greeter-header-search">
        <div className="greeter-search-full">
          <FullSearchTrigger hideIfDisabled />
        </div>
        <SearchTrigger hideIfDisabled aria-label="Open search" className="greeter-search-icon" />
      </div>
      <div className="greeter-header-meta">
        <a
          className="greeter-header-badge"
          href="https://github.com/Nomadcxx/sysc-greet/stargazers"
          aria-label="sysc-greet GitHub stars"
        >
          <img
            src="https://img.shields.io/github/stars/Nomadcxx/sysc-greet?style=flat-square&logo=github&label=stars&color=ef233c&labelColor=2b2d42"
            alt="GitHub stars"
            width="85"
            height="20"
          />
        </a>
        <a
          className="greeter-header-badge"
          href="https://github.com/Nomadcxx/sysc-greet/blob/master/go.mod"
          aria-label="sysc-greet uses Go 1.25.1"
        >
          <img
            src="https://img.shields.io/badge/Go-1.25.1-00ADD8?style=flat-square&logo=go&logoColor=white&labelColor=2b2d42"
            alt="Go 1.25.1"
            width="87"
            height="20"
          />
        </a>
        <SidebarTrigger className="greeter-sidebar-trigger" aria-label="Open navigation">
          <PanelLeft aria-hidden="true" />
        </SidebarTrigger>
      </div>
    </header>
  );
}
