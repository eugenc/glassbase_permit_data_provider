import { useLocation } from 'react-router-dom';

const titles: Record<string, string> = {
  '/': 'Dashboard',
  '/counties': 'Counties',
  '/runs': 'Run History',
};

export default function TopBar() {
  const loc = useLocation();
  const title =
    titles[loc.pathname] ??
    (loc.pathname.startsWith('/counties') ? 'County' : 'GlassBase');

  return (
    <header className="fixed left-[240px] right-0 top-0 z-10 flex h-16 items-center border-b border-border bg-canvas/95 px-8 backdrop-blur">
      <h1 className="text-xl font-semibold text-text-1">{title}</h1>
    </header>
  );
}
