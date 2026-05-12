import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Navigate, Route, Routes } from 'react-router-dom';
import { AuthProvider, useAuth } from './auth/AuthContext';
import Shell from './components/layout/Shell';
import { ToastProvider } from './components/ui/ToastProvider';
import CountyDetail from './pages/CountyDetail';
import Counties from './pages/Counties';
import Dashboard from './pages/Dashboard';
import Login from './pages/Login';
import Permits from './pages/Permits';
import Repairs from './pages/Repairs';
import RunHistory from './pages/RunHistory';

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, retry: 1 } },
});

function ProtectedRoutes() {
  const { user, isLoading } = useAuth();
  if (isLoading) return <div className="min-h-screen bg-canvas" />;
  if (!user) return <Navigate to="/login" replace />;
  return (
    <Shell>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/counties" element={<Counties />} />
        <Route path="/counties/:id" element={<CountyDetail />} />
        <Route path="/counties/:id/permits" element={<Permits />} />
        <Route path="/runs" element={<RunHistory />} />
        <Route path="/repairs" element={<Repairs />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Shell>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <AuthProvider>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/*" element={<ProtectedRoutes />} />
          </Routes>
        </AuthProvider>
      </ToastProvider>
    </QueryClientProvider>
  );
}
