import { NavLink } from 'react-router-dom';
import { useAuth } from '../../auth/AuthContext';

const navItems = [
  { path: '/', label: 'Dashboard' },
  { path: '/counties', label: 'Counties' },
  { path: '/runs', label: 'Run History' },
];

export default function Sidebar() {
  const { user, logout } = useAuth();

  return (
    <aside className="fixed left-0 top-0 z-20 flex h-screen w-[240px] flex-col border-r border-border bg-canvas">
      <div className="flex h-16 items-center border-b border-border px-6 py-5">
        <img src="/logo-white.svg" alt="GlassBase" className="h-[18px] w-auto" />
      </div>
      <nav className="flex-1 space-y-0.5 px-2 py-4">
        {navItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            end={item.path === '/'}
            className={({ isActive }) =>
              `my-0.5 flex items-center rounded-md px-4 py-2.5 text-[13px] font-medium transition ` +
              (isActive
                ? 'bg-brand font-semibold text-white'
                : 'text-text-2 hover:bg-surface hover:text-text-1')
            }
          >
            {item.label}
          </NavLink>
        ))}
      </nav>
      <div className="border-t border-border p-4">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-border text-xs font-medium text-brand">
            {user?.name?.[0]?.toUpperCase()}
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-[13px] font-medium text-text-1">{user?.name}</p>
            <p className="text-[11px] text-text-2">{user?.role}</p>
          </div>
          <button
            type="button"
            onClick={logout}
            className="text-[11px] text-text-3 hover:text-text-1 transition"
          >
            Log out
          </button>
        </div>
      </div>
    </aside>
  );
}
