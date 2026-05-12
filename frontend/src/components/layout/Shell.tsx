import GlobalLoader from '../ui/GlobalLoader';
import Sidebar from './Sidebar';
import TopBar from './TopBar';

export default function Shell({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-canvas">
      <GlobalLoader />
      <Sidebar />
      <TopBar />
      <main className="ml-[240px] min-h-screen pt-16">
        <div className="mx-auto max-w-[1200px] px-8 py-8">{children}</div>
      </main>
    </div>
  );
}
