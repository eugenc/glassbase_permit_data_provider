import type { Permit } from '../../types';

interface Props {
  permit: Permit | null;
  onClose: () => void;
}

export default function PermitDrawer({ permit, onClose }: Props) {
  if (!permit) return null;

  const rawData = permit.raw_data ?? {};

  return (
    <>
      <button
        type="button"
        aria-label="Close drawer backdrop"
        className="fixed inset-0 z-40 bg-black/40"
        onClick={onClose}
      />
      <div className="fixed right-0 top-0 z-50 flex h-screen w-96 flex-col border-l border-border bg-surface shadow-modal">
        <div className="flex items-center justify-between border-b border-border px-5 py-4">
          <div>
            <p className="text-xs text-text-3">Permit</p>
            <p className="font-mono text-sm font-medium text-text-1">{permit.permit_number}</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-xl text-text-3 transition hover:text-text-1"
          >
            ✕
          </button>
        </div>

        <div className="flex gap-4 border-b border-border px-5 py-3">
          <div>
            <p className="text-xs text-text-3">Scraped</p>
            <p className="text-xs text-text-1">{new Date(permit.scraped_at).toLocaleDateString()}</p>
          </div>
        </div>

        <div className="flex-1 space-y-2 overflow-y-auto px-5 py-4">
          {Object.entries(rawData).map(([key, value]) => (
            <div key={key} className="flex flex-col gap-0.5">
              <span className="text-[11px] font-medium uppercase tracking-wide text-text-3">
                {key.replace(/_/g, ' ')}
              </span>
              <span className="break-words font-mono text-sm text-text-1">
                {value === null || value === undefined || value === '' ? (
                  <span className="text-text-3">—</span>
                ) : typeof value === 'object' ? (
                  JSON.stringify(value)
                ) : (
                  String(value)
                )}
              </span>
            </div>
          ))}
        </div>

        <div className="border-t border-border px-5 py-3">
          <p className="font-mono text-xs text-text-3">ID: {permit.id}</p>
        </div>
      </div>
    </>
  );
}
