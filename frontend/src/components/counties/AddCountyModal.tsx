import { useState } from 'react';
import axios from 'axios';
import Modal from '../ui/Modal';
import Button from '../ui/Button';
import { useAddCounty } from '../../hooks/useCounties';

interface Props {
  onClose: () => void;
}

type Stage = 'form' | 'processing' | 'success' | 'error';

export default function AddCountyModal({ onClose }: Props) {
  const addCounty = useAddCounty();
  const [stage, setStage] = useState<Stage>('form');
  const [error, setError] = useState('');

  const [form, setForm] = useState({
    county_name: '',
    state: '',
    url: '',
  });

  const countyId = `${form.county_name.toLowerCase().replace(/\s+/g, '_')}_${form.state.toLowerCase()}`;

  const parseAxiosMessage = (err: unknown) => {
    if (!axios.isAxiosError(err)) return 'Something went wrong';
    const d = err.response?.data;
    if (typeof d === 'string') return d;
    return err.message || 'Request failed';
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setStage('processing');
    try {
      await addCounty.mutateAsync({
        county_id: countyId,
        county_name: form.county_name,
        state: form.state.toUpperCase(),
        url: form.url,
      });
      setStage('success');
    } catch (err: unknown) {
      setError(parseAxiosMessage(err));
      setStage('error');
    }
  };

  if (stage === 'processing') {
    return (
      <Modal title="Adding County" onClose={() => {}}>
        <div className="space-y-4 py-6 text-center">
          <div className="mx-auto h-10 w-10 animate-spin rounded-full border-2 border-brand border-t-transparent" />
          <div>
            <p className="text-sm font-medium text-text-1">Analyzing permit page...</p>
            <p className="mt-1 text-xs text-text-3">
              Claude is reading the page structure. This may take 15–30 seconds.
            </p>
          </div>
          <div className="space-y-1.5 rounded-lg bg-canvas p-3 text-left font-mono text-xs text-text-3">
            <p>→ Fetching {form.url}</p>
            <p>→ Detecting source type...</p>
            <p>→ Generating connector config...</p>
            <p>→ Writing to registry...</p>
          </div>
        </div>
      </Modal>
    );
  }

  if (stage === 'success') {
    return (
      <Modal title="County Added" onClose={onClose}>
        <div className="space-y-4 py-6 text-center">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-status-active-bg text-2xl text-status-active">
            ✓
          </div>
          <div>
            <p className="text-sm font-medium text-text-1">{form.county_name} is ready</p>
            <p className="mt-1 font-mono text-xs text-text-3">{countyId}</p>
          </div>
          <p className="text-xs text-text-3">
            The connector has been generated and the county is active. It will run with scheduled scrapes.
          </p>
          <Button variant="secondary" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>
      </Modal>
    );
  }

  if (stage === 'error') {
    return (
      <Modal title="Failed to Add County" onClose={onClose}>
        <div className="space-y-4 py-4">
          <div className="rounded-lg border border-status-broken bg-status-broken-bg p-3">
            <p className="text-xs text-status-broken">{error}</p>
          </div>
          <div className="flex gap-2">
            <Button variant="secondary" size="sm" onClick={() => setStage('form')}>
              Try Again
            </Button>
            <Button variant="ghost" size="sm" onClick={onClose}>
              Cancel
            </Button>
          </div>
        </div>
      </Modal>
    );
  }

  return (
    <Modal title="Add County" onClose={onClose}>
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-4">
        <div className="grid grid-cols-2 gap-3">
          <div className="col-span-2">
            <label className="mb-1.5 block text-xs font-medium text-text-2">County Name</label>
            <input
              required
              value={form.county_name}
              onChange={(e) => setForm((f) => ({ ...f, county_name: e.target.value }))}
              placeholder="Broward County"
              className="w-full rounded-md border border-border bg-surface px-3.5 py-2 text-sm text-text-1 placeholder:text-text-3 focus:border-brand focus:outline-none"
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium text-text-2">State</label>
            <input
              required
              maxLength={2}
              value={form.state}
              onChange={(e) =>
                setForm((f) => ({ ...f, state: e.target.value.toUpperCase() }))
              }
              placeholder="FL"
              className="w-full rounded-md border border-border bg-surface px-3.5 py-2 text-sm uppercase text-text-1 placeholder:text-text-3 focus:border-brand focus:outline-none"
            />
          </div>
          <div className="col-span-2">
            <label className="mb-1.5 block text-xs font-medium text-text-2">Permit Search URL</label>
            <input
              required
              type="url"
              value={form.url}
              onChange={(e) => setForm((f) => ({ ...f, url: e.target.value }))}
              placeholder="https://permits.county.gov/search"
              className="w-full rounded-md border border-border bg-surface px-3.5 py-2 text-sm text-text-1 placeholder:text-text-3 focus:border-brand focus:outline-none"
            />
          </div>
        </div>

        {form.county_name && form.state && (
          <p className="text-xs text-text-3">
            ID: <span className="font-mono text-text-2">{countyId}</span>
          </p>
        )}

        <p className="text-xs text-text-3">
          Claude analyzes the permit page and generates a scraping connector. Expect 15–30 seconds.
        </p>

        <div className="flex gap-2 pt-1">
          <Button type="submit" size="sm">
            Add County
          </Button>
          <Button type="button" variant="ghost" size="sm" onClick={onClose}>
            Cancel
          </Button>
        </div>
      </form>
    </Modal>
  );
}
