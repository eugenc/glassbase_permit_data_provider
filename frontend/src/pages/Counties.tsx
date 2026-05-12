import { useMemo, useState } from 'react';
import AddCountyModal from '../components/counties/AddCountyModal';
import CountyRow from '../components/counties/CountyRow';
import Button from '../components/ui/Button';
import EmptyState from '../components/ui/EmptyState';
import { useCounties } from '../hooks/useCounties';

const US_STATES = [
  'All',
  'AL',
  'AK',
  'AZ',
  'AR',
  'CA',
  'CO',
  'CT',
  'DE',
  'FL',
  'GA',
  'HI',
  'ID',
  'IL',
  'IN',
  'IA',
  'KS',
  'KY',
  'LA',
  'ME',
  'MD',
  'MA',
  'MI',
  'MN',
  'MS',
  'MO',
  'MT',
  'NE',
  'NV',
  'NH',
  'NJ',
  'NM',
  'NY',
  'NC',
  'ND',
  'OH',
  'OK',
  'OR',
  'PA',
  'RI',
  'SC',
  'SD',
  'TN',
  'TX',
  'UT',
  'VT',
  'VA',
  'WA',
  'WV',
  'WI',
  'WY',
];

export default function Counties() {
  const { data, isLoading } = useCounties();
  const [showAdd, setShowAdd] = useState(false);
  const [search, setSearch] = useState('');
  const [stateFilter, setStateFilter] = useState('All');
  const [statusFilter, setStatusFilter] = useState('All');

  const counties = useMemo(() => {
    if (!data?.counties) return [];
    return data.counties.filter((c) => {
      const matchSearch =
        !search ||
        c.county_name.toLowerCase().includes(search.toLowerCase()) ||
        c.county_id.includes(search.toLowerCase());
      const matchState = stateFilter === 'All' || c.state === stateFilter;
      const matchStatus = statusFilter === 'All' || c.status === statusFilter;
      return matchSearch && matchState && matchStatus;
    });
  }, [data, search, stateFilter, statusFilter]);

  const emptyRegistry = !isLoading && data?.counties?.length === 0;

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <p className="mt-0.5 text-sm text-text-2">{data?.total ?? '—'} counties configured</p>
        </div>
        <Button onClick={() => setShowAdd(true)}>+ Add County</Button>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search counties..."
          className="w-60 rounded-md border border-border bg-surface px-3.5 py-2 text-sm text-text-1 placeholder:text-text-3 focus:border-brand focus:outline-none"
        />
        <select
          value={stateFilter}
          onChange={(e) => setStateFilter(e.target.value)}
          className="rounded-md border border-border bg-surface px-3 py-2 text-sm text-text-1 focus:border-brand focus:outline-none"
        >
          {US_STATES.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-md border border-border bg-surface px-3 py-2 text-sm text-text-1 focus:border-brand focus:outline-none"
        >
          {['All', 'active', 'broken', 'paused'].map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
        <span className="ml-auto text-xs text-text-3">{counties.length} shown</span>
      </div>

      <div className="overflow-hidden rounded-[10px] border border-border bg-surface shadow-card">
        {emptyRegistry ? (
          <EmptyState
            icon="◫"
            title="No counties yet"
            description="Add your first county to start collecting permit data."
            action={{ label: '+ Add County', onClick: () => setShowAdd(true) }}
          />
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-border bg-[#0D1220]">
                {['County', 'State', 'Status', 'Permits', 'Last Run', 'Type', 'Actions'].map((h) => (
                  <th
                    key={h}
                    className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.06em] text-text-3"
                  >
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {isLoading
                ? [...Array(8)].map((_, i) => <SkeletonRow key={i} />)
                : counties.map((c) => <CountyRow key={c.county_id} county={c} />)}
            </tbody>
          </table>
        )}

        {!isLoading && !emptyRegistry && counties.length === 0 && (
          <div className="py-12 text-center text-sm text-text-3">No counties match your filters.</div>
        )}
      </div>

      {showAdd && <AddCountyModal onClose={() => setShowAdd(false)} />}
    </div>
  );
}

function SkeletonRow() {
  return (
    <tr className="animate-pulse border-b border-border">
      {[...Array(7)].map((_, i) => (
        <td key={i} className="px-4 py-4">
          <div className="h-4 max-w-[75%] rounded bg-border" />
        </td>
      ))}
    </tr>
  );
}
