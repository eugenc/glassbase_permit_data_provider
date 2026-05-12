import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

export default function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await login(email, password);
      navigate('/');
    } catch {
      setError('Invalid email or password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-canvas px-4">
      <div className="w-full max-w-sm">
        <div className="mb-10 text-center">
          <img src="/logo-white.svg" alt="GlassBase" className="mx-auto mb-4 h-8 w-auto" />
          <p className="text-sm text-text-3">Permit Intelligence Platform</p>
        </div>

        <form
          onSubmit={(e) => void handleSubmit(e)}
          className="space-y-5 rounded-xl border border-border bg-surface p-8 shadow-card"
        >
          <div>
            <label htmlFor="email" className="mb-1.5 block text-xs font-medium text-text-2">
              Email
            </label>
            <input
              id="email"
              type="email"
              autoComplete="username"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full rounded-md border border-border bg-canvas px-3.5 py-2.5 text-sm text-text-1 placeholder:text-text-3 focus:border-brand focus:outline-none"
              placeholder="you@glassbase.io"
              required
            />
          </div>
          <div>
            <label htmlFor="password" className="mb-1.5 block text-xs font-medium text-text-2">
              Password
            </label>
            <input
              id="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full rounded-md border border-border bg-canvas px-3.5 py-2.5 text-sm text-text-1 focus:border-brand focus:outline-none"
              required
            />
          </div>

          {error && <p className="text-sm text-status-broken">{error}</p>}

          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-md bg-brand py-2.5 text-sm font-semibold text-white transition hover:bg-brand-hover disabled:opacity-50"
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  );
}
