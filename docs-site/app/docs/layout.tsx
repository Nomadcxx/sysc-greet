import { source } from '@/lib/source';
import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import { baseOptions } from '@/lib/layout.shared';
import { OperatorHeader } from '@/components/operator-header';

export default function Layout({ children }: LayoutProps<'/docs'>) {
  return (
    <DocsLayout
      tree={source.getPageTree()}
      {...baseOptions()}
      containerProps={{ className: 'docs-shell' }}
      slots={{ header: OperatorHeader }}
    >
      {children}
    </DocsLayout>
  );
}
