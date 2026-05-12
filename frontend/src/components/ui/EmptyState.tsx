interface Props {
  icon?: string;
  title: string;
  description?: string;
  action?: { label: string; onClick: () => void };
}

export default function EmptyState({ icon = '◻', title, description, action }: Props) {
  return (
    <div className="py-16 text-center">
      <div className="mb-4 text-4xl opacity-30">{icon}</div>
      <p className="text-sm font-medium text-text-2">{title}</p>
      {description && (
        <p className="mx-auto mt-1 max-w-xs text-xs text-text-3">{description}</p>
      )}
      {action && (
        <button
          type="button"
          onClick={action.onClick}
          className="mt-4 text-xs font-medium text-brand hover:text-brand-hover transition"
        >
          {action.label}
        </button>
      )}
    </div>
  );
}
