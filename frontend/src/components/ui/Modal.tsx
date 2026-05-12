import { useEffect } from 'react';

interface Props {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
}

export default function Modal({ title, onClose, children }: Props) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <button
        type="button"
        aria-label="Close modal backdrop"
        className="absolute inset-0 bg-black/65 backdrop-blur-sm"
        onClick={onClose}
      />
      <div className="relative z-10 w-full max-w-md rounded-xl border border-border bg-surface p-7 shadow-modal">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-base font-semibold text-text-1">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            className="text-xl leading-none text-text-3 transition hover:text-text-1"
          >
            ✕
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}
