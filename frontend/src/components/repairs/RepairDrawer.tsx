import { useQueryClient } from '@tanstack/react-query';
import { createPortal } from 'react-dom';
import { useEffect, useRef, useState } from 'react';
import Button from '../ui/Button';

interface Props {
  countyId: string;
  countyName: string;
  onClose: () => void;
}

type LineStatus = 'connecting' | 'running' | 'success' | 'failed';

type LogEntry =
  | { text: string; stream: 'stdout' | 'stderr' }
  | { text: string; stream: 'pulse' };

/** Parse one SSE frame (possibly multiple `data:` lines joined with newlines per spec). */
function parseSSEFrame(chunk: string): { eventType: string; dataRaw: string } | null {
  let eventType = '';
  const dataLines: string[] = [];

  for (const raw of chunk.split('\n')) {
    if (raw.startsWith('event: ')) eventType = raw.slice(7).trim();
    else if (raw.startsWith('event:')) eventType = raw.slice(6).trim();
    else if (raw.startsWith('data: ')) dataLines.push(raw.slice(6));
    else if (raw.startsWith('data:')) dataLines.push(raw.slice(5));
  }

  if (!eventType || dataLines.length === 0) return null;
  return { eventType, dataRaw: dataLines.join('\n') };
}

export default function RepairDrawer({ countyId, countyName, onClose }: Props) {
  const [lines, setLines] = useState<LogEntry[]>([]);
  const [status, setStatus] = useState<LineStatus>('connecting');
  const [result, setResult] = useState<{ commitSha?: string; prUrl?: string } | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const qc = useQueryClient();

  useEffect(() => {
    const token = localStorage.getItem('gb_token');
    const controller = new AbortController();
    const base = import.meta.env.VITE_API_URL ?? '';

    function applyParsedEvent(eventType: string, data: Record<string, unknown>) {
      if (eventType === 'start') {
        setStatus('running');
        return;
      }
      if (eventType === 'pulse' && typeof data.message === 'string') {
        setLines((prev) => {
          const next = [...prev];
          let pulseIdx = -1;
          for (let j = next.length - 1; j >= 0; j--) {
            if (next[j].stream === 'pulse') {
              pulseIdx = j;
              break;
            }
          }
          const entry = { text: data.message as string, stream: 'pulse' as const };
          if (pulseIdx >= 0) next[pulseIdx] = entry;
          else next.push(entry);
          return next;
        });
        return;
      }
      if (eventType === 'output' && typeof data.line === 'string') {
        const stream = data.stream === 'stderr' ? 'stderr' : ('stdout' as const);
        const lineStr: string = data.line;
        setLines((prev) => {
          const withoutPulse = prev.filter((e) => e.stream !== 'pulse');
          return [...withoutPulse, { text: lineStr, stream }];
        });
        return;
      }
      if (eventType === 'error' && typeof data.message === 'string') {
        setLines((prev) => [
          ...prev.filter((e) => e.stream !== 'pulse'),
          { text: `Error — ${data.message}`, stream: 'stderr' },
        ]);
        setStatus('failed');
        return;
      }
      if (eventType === 'complete') {
        const ok = Boolean(data.success);
        if (ok) {
          setStatus('success');
          setResult({
            commitSha: typeof data.commit_sha === 'string' ? data.commit_sha : undefined,
            prUrl: typeof data.pr_url === 'string' ? data.pr_url : undefined,
          });
        } else {
          setStatus('failed');
        }
        const msg =
          typeof data.message === 'string'
            ? data.message
            : 'Claude Code could not repair this county.';
        setLines((prev) => {
          let base = prev.filter((e) => e.stream !== 'pulse');
          if (ok) base = [...base, { text: 'Repair complete', stream: 'stdout' as const }];
          else base = [...base, { text: msg, stream: 'stderr' as const }];
          return base;
        });
        void qc.invalidateQueries({ queryKey: ['counties'] });
        void qc.invalidateQueries({ queryKey: ['repairs'] });
        void qc.invalidateQueries({ queryKey: ['county', countyId] });
      }
    }

    async function stream() {
      const resp = await fetch(`${base}/api/counties/${countyId}/repair-cc`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token ?? ''}`,
          Accept: 'text/event-stream',
        },
        signal: controller.signal,
      });

      if (!resp.ok || !resp.body) {
        setStatus('failed');
        setLines((prev) => [...prev, { text: `Error ${resp.status}: ${resp.statusText}`, stream: 'stderr' }]);
        return;
      }

      const reader = resp.body.getReader();
      const decoder = new TextDecoder();
      let buf = '';

      function flushChunks(raw: string) {
        const parts = raw.split(/\n\n/);
        const tail = parts.pop() ?? '';
        for (const evt of parts) {
          if (!evt.trim()) continue;
          const parsed = parseSSEFrame(evt);
          if (!parsed) continue;
          try {
            const data = JSON.parse(parsed.dataRaw) as Record<string, unknown>;
            applyParsedEvent(parsed.eventType, data);
          } catch {
            /* ignore malformed JSON */
          }
        }
        return tail;
      }

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buf += decoder.decode(value, { stream: true });
        buf = flushChunks(buf);
      }
      buf += decoder.decode();
      buf = flushChunks(buf + '\n\n');
    }

    void stream().catch((e: Error) => {
      if (!controller.signal.aborted) {
        setStatus('failed');
        setLines((prev) => [...prev, { text: String(e), stream: 'stderr' }]);
      }
    });

    return () => controller.abort();
  }, [countyId, qc]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [lines]);

  const labels: Record<LineStatus, string> = {
    connecting: 'Connecting…',
    running: 'Claude Code running…',
    success: 'Complete',
    failed: 'Failed',
  };

  return createPortal(
    <>
      <div className="fixed inset-0 z-40 bg-black/50" aria-hidden="true" onClick={onClose} />
      <div className="fixed right-0 top-0 z-50 flex h-screen w-[min(560px,100vw)] flex-col border-l border-border bg-canvas shadow-2xl">
        <header className="flex items-center justify-between border-b border-border px-5 py-4">
          <div>
            <p className="text-[11px] font-semibold uppercase tracking-wider text-text-3">
              Claude Code repair
            </p>
            <p className="text-sm font-medium text-text-1">{countyName}</p>
          </div>
          <div className="flex items-center gap-3">
            <span
              className={
                status === 'failed'
                  ? 'text-status-broken text-xs font-medium'
                  : status === 'success'
                    ? 'text-status-active text-xs font-medium'
                    : 'text-text-3 text-xs font-medium'
              }
            >
              {status === 'running' ? (
                <span className="mr-2 inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-brand" />
              ) : null}
              {labels[status]}
            </span>
            <button
              type="button"
              onClick={onClose}
              className="text-sm text-text-3 hover:text-text-1"
              aria-label="Close"
            >
              ✕
            </button>
          </div>
        </header>

        <div className="scrollbar-thin flex min-h-0 flex-1 flex-col bg-elevated/30">
          <div className="min-h-0 flex-1 overflow-y-auto p-4 font-mono text-xs leading-relaxed">
            {lines.length === 0 && status === 'connecting' ? (
              <p className="animate-pulse text-text-3">Launching Claude Code…</p>
            ) : null}
            {lines.map((entry, i) => (
              <div
                key={i}
                className={
                  entry.stream === 'pulse'
                    ? 'whitespace-pre-wrap rounded-md border border-border bg-elevated/80 px-2 py-1.5 text-[11px] italic text-text-3'
                    : entry.stream === 'stderr'
                      ? 'whitespace-pre-wrap text-text-3 opacity-90'
                      : 'whitespace-pre-wrap text-text-2'
                }
              >
                {entry.text.trim() === '' ? '\u00A0' : entry.text}
              </div>
            ))}
            <div ref={bottomRef} />
          </div>
        </div>

        {(status === 'success' || status === 'failed') && (
          <footer className="space-y-2 border-t border-border px-5 py-4">
            {status === 'success' && result?.commitSha ? (
              <p className="text-xs text-text-3">
                Commit <span className="font-mono text-text-1">{result.commitSha}</span>
              </p>
            ) : null}
            {status === 'success' && result?.prUrl ? (
              <a
                href={result.prUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs font-medium text-brand hover:underline"
              >
                Open pull request
              </a>
            ) : null}
            {status === 'failed' ? (
              <p className="text-xs text-text-3">
                County marked paused unless repair succeeded elsewhere. Inspect output above for
                detail.
              </p>
            ) : null}
            <Button type="button" variant="secondary" size="sm" className="w-full" onClick={onClose}>
              Close
            </Button>
          </footer>
        )}
      </div>
    </>,
    document.body,
  );
}
