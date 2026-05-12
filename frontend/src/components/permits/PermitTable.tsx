import type { Permit } from '../../types';

const SKIP_COLS = new Set(['id', 'raw_data']);
const MAX_DISPLAY_COLS = 6;

interface Props {
  permits: Permit[];
  onSelect: (permit: Permit) => void;
  selected: Permit | null;
}

export default function PermitTable({ permits, onSelect, selected }: Props) {
  if (!permits.length) {
    return <div className="py-12 text-center text-sm text-text-3">No permits found.</div>;
  }

  const allCols = Object.keys(permits[0]).filter((k) => !SKIP_COLS.has(k));
  const priority = [
    'permit_number',
    'scraped_at',
    'permit_type',
    'address',
    'status',
    'issue_date',
    'contractor',
  ];
  const cols = [
    ...priority.filter((p) => allCols.includes(p)),
    ...allCols.filter((c) => !priority.includes(c)),
  ].slice(0, MAX_DISPLAY_COLS);

  const formatCol = (col: string) =>
    col.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase());

  const formatVal = (key: string, val: unknown) => {
    if (key === 'scraped_at' && val) {
      return new Date(String(val)).toLocaleDateString();
    }
    if (val === null || val === undefined || val === '') return '—';
    return String(val);
  };

  return (
    <table className="w-full">
      <thead>
        <tr className="border-b border-border bg-[#0D1220]">
          {cols.map((col) => (
            <th
              key={col}
              className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.06em] text-text-3"
            >
              {formatCol(col)}
            </th>
          ))}
          <th className="px-4 py-3" />
        </tr>
      </thead>
      <tbody>
        {permits.map((permit) => (
          <tr
            key={permit.id}
            onClick={() => onSelect(permit)}
            className={`cursor-pointer border-b border-border transition hover:bg-[#151D30] ${
              selected?.id === permit.id ? 'border-brand bg-brand/10' : ''
            }`}
          >
            {cols.map((col) => (
              <td key={col} className="px-4 py-3">
                {col === 'permit_number' ? (
                  <span className="font-mono text-xs text-brand">{formatVal(col, permit[col])}</span>
                ) : (
                  <span className="text-sm text-text-2">{formatVal(col, permit[col])}</span>
                )}
              </td>
            ))}
            <td className="px-4 py-3 text-xs text-text-3">→</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
