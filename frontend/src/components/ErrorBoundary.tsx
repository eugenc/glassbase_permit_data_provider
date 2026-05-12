import { Component, type ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: string;
}

export default class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: '' };
  }

  static getDerivedStateFromError(err: Error) {
    return { hasError: true, error: err.message };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex min-h-screen items-center justify-center bg-canvas">
          <div className="max-w-md text-center">
            <p className="mb-4 text-4xl">⚠</p>
            <h2 className="mb-2 text-lg font-semibold text-text-1">Something went wrong</h2>
            <p className="mb-6 font-mono text-sm text-text-2">{this.state.error}</p>
            <button
              type="button"
              onClick={() => {
                window.location.href = '/';
              }}
              className="rounded-md bg-brand px-4 py-2 text-sm font-semibold text-white transition hover:bg-brand-hover"
            >
              Return to Dashboard
            </button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
