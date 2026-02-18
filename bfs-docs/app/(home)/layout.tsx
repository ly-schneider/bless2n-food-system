import { source } from '@/lib/source';
import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import { baseOptions } from '@/lib/layout.shared';

const version = process.env.NEXT_PUBLIC_APP_VERSION;

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <DocsLayout
      tree={source.pageTree}
      {...baseOptions()}
      sidebar={{
        footer: version ? <span className="text-xs text-gray-400 text-center">v{version}</span> : null,
      }}
    >
      {children}
    </DocsLayout>
  );
}
