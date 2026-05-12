import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { downloadPermitsCsv } from '../api/client';
import PermitDrawer from '../components/permits/PermitDrawer';
import PermitTable from '../components/permits/PermitTable';
import Badge from '../components/ui/Badge';
import Button from '../components/ui/Button';
import { useCounties } from '../hooks/useCounties';
import { usePermits } from '../hooks/usePermits';
import type { Permit } from '../types';

const PER_PAGE = 50;

export default function Permits() {
  const { id } = useParams<{ id: string }>();
  const countyId = id ?? '';
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [selected, setSelected] = useState<Permit | null>(null);

  const { data: countiesData } = useCounties();
  const county = countiesData?.counties.find((c) => c.county_id === countyId);

  const { data, isLoading, isFetching } = usePermits(countyId, {
    page,
    per_page: PER_PAGE,
    search,
  });

  const totalPages = data ? Math.ceil(data.total / PER_PAGE) : 0;

  const handleExport = async () => {
    try {
      await downloadPermitsCsv(countyId);
    } catch {
      /* handled globally */
    }
  };

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="mb-1 flex items-center gap-2">
            <Link to="/counties" className="text-xs text-text-3 transition hover:text-text-1">
              Counties
            </Link>
            <span className="text-border">/</span>
            <span className="text-xs text-text-2">{county?.county_name ?? countyId}</span>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <h2 className="text-xl font-semibold text-text-1">Permits</h2>
            {county && <Badge status={county.status} />}
          </div>
          <p className="mt-0.5 text-sm text-text-3">
            <span className="font-mono">{data?.total?.toLocaleString() ?? '—'}</span> total records
            {county && <span className="ml-2 text-text-3">· {county.source_type}</span>}
          </p>
        </div>
        <Button variant="secondary" size="sm" onClick={() => void handleExport()}>
          ↓ Export CSV
        </Button>
      </div>

      <input
        value={search}
        onChange={(e) => {
          setSearch(e.target.value);
          setPage(1);
        }}
        placeholder="Search by permit number, address..."
        className="w-full max-w-md rounded-md border border-border bg-surface px-3.5 py-2 text-sm text-text-1 placeholder:text-text-3 focus:border-brand focus:outline-none md:w-96"
      />

      <div
        className={`overflow-hidden rounded-[10px] border border-border bg-surface shadow-card transition ${isFetching ? 'opacity-70' : ''}`}
      >
        {isLoading ? (
          <div className="py-12 text-center text-text-3">Loading...</div>
        ) : (
          <PermitTable
            permits={(data?.permits as Permit[]) ?? []}
            onSelect={setSelected}
            selected={selected}
          />
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-xs text-text-3">
            Page {page} of {totalPages}
          </p>
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="secondary"
              disabled={page === 1}
              onClick={() => setPage((p) => p - 1)}
            >
              ← Prev
            </Button>
            <Button
              size="sm"
              variant="secondary"
              disabled={page === totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              Next →
            </Button>
          </div>
        </div>
      )}

      <PermitDrawer permit={selected} onClose={() => setSelected(null)} />
    </div>
  );
}
