import { Provider } from '@/components/provider';
import type { Metadata } from 'next';
import './global.css';

export const metadata: Metadata = {
  metadataBase: new URL('https://nomadcxx.github.io'),
  title: {
    default: 'sysc-greet documentation',
    template: '%s | sysc-greet',
  },
  description: 'Documentation for the sysc-greet graphical console greeter.',
  icons: { icon: `${process.env.NEXT_PUBLIC_BASE_PATH ?? ''}/logo.png` },
};

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <html lang="en" className="dark" style={{ colorScheme: 'dark' }}>
      <body className="flex flex-col min-h-screen">
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}
