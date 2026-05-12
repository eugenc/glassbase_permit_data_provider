type Variant = 'primary' | 'secondary' | 'danger' | 'ghost';
type Size = 'sm' | 'md';

interface Props extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
}

const variants: Record<Variant, string> = {
  primary:
    'bg-brand hover:bg-brand-hover text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2 focus-visible:ring-offset-canvas',
  secondary:
    'border border-brand bg-transparent text-brand hover:bg-brand-light/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand',
  danger:
    'border border-status-broken bg-status-broken-bg text-status-broken hover:bg-status-broken-bg/80',
  ghost: 'text-text-2 hover:text-text-1',
};

const sizes: Record<Size, string> = {
  sm: 'text-xs px-3 py-1.5 rounded-md',
  md: 'text-sm px-4 py-2.5 rounded-md font-semibold',
};

export default function Button({
  variant = 'primary',
  size = 'md',
  loading,
  children,
  className = '',
  disabled,
  type = 'button',
  ...props
}: Props) {
  return (
    <button
      {...props}
      type={type}
      disabled={loading || disabled}
      className={`inline-flex items-center justify-center gap-2 font-medium transition disabled:cursor-not-allowed disabled:opacity-50 ${variants[variant]} ${sizes[size]} ${className}`}
    >
      {loading && (
        <span className="h-3 w-3 animate-spin rounded-full border border-current border-t-transparent" />
      )}
      {children}
    </button>
  );
}
