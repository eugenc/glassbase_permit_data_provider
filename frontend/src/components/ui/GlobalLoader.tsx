import { useIsFetching, useIsMutating } from '@tanstack/react-query';

export default function GlobalLoader() {
  const fetching = useIsFetching();
  const mutating = useIsMutating();
  const active = fetching + mutating > 0;

  return (
    <div
      className={`fixed left-0 right-0 top-0 z-[200] h-0.5 transition-opacity duration-300 ${active ? 'opacity-100' : 'opacity-0'}`}
    >
      <div className="h-full animate-pulse bg-brand" />
    </div>
  );
}
