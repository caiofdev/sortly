function OrganizerActions({
  labels,
  isLoading,
  loadingAction,
  hasUndo,
  hasSource,
  onOrganizeFiles,
  onUndoLastOrganization
}) {
  const isOrganizing = loadingAction === 'organize';
  const isRestoring = loadingAction === 'restore';
  const baseIconButtonClass =
    'group relative inline-flex h-12 w-12 items-center justify-center rounded-xl transition-all';

  return (
    <div className="flex items-center justify-center gap-3">
      <button
        type="button"
        onClick={onOrganizeFiles}
        disabled={isLoading || !hasSource}
        title={isOrganizing ? labels.organizing : labels.organize}
        aria-label={isOrganizing ? labels.organizing : labels.organize}
            className={`${baseIconButtonClass} bg-[#22C55E] text-white hover:-translate-y-0.5 hover:bg-[#32d26b] disabled:cursor-not-allowed disabled:bg-[#1c4a33] disabled:text-slate-300`}
      >
        {isOrganizing ? (
          <svg viewBox="0 0 24 24" className="h-5 w-5 animate-spin" aria-hidden="true">
            <circle cx="12" cy="12" r="9" className="opacity-25" stroke="currentColor" strokeWidth="3" fill="none" />
            <path className="opacity-90" fill="currentColor" d="M12 3a9 9 0 0 1 9 9h-3a6 6 0 0 0-6-6V3Z" />
          </svg>
        ) : (
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="h-5 w-5" aria-hidden="true">
            <path d="M8 5v14l11-7-11-7Z" fill="currentColor" stroke="none" />
            <path d="M4 5v14" />
          </svg>
        )}
        <span className="pointer-events-none absolute -bottom-8 left-1/2 z-10 w-max -translate-x-1/2 rounded-md bg-[#0F172A] px-2 py-1 text-xs text-[#94A3B8] opacity-0 shadow-md transition-opacity group-hover:opacity-100">
          {isOrganizing ? labels.organizing : labels.organize}
        </span>
      </button>

      <button
        type="button"
        onClick={onUndoLastOrganization}
        disabled={isLoading || !hasUndo}
        title={labels.undo}
        aria-label={labels.undo}
        className={`${baseIconButtonClass} bg-[#DC2626] text-white hover:-translate-y-0.5 hover:bg-[#ef4444] disabled:cursor-not-allowed disabled:bg-[#4f2020] disabled:text-slate-300`}
      >
        {isRestoring ? (
          <svg viewBox="0 0 24 24" className="h-5 w-5 animate-spin" aria-hidden="true">
            <circle cx="12" cy="12" r="9" className="opacity-25" stroke="currentColor" strokeWidth="3" fill="none" />
            <path className="opacity-90" fill="currentColor" d="M12 3a9 9 0 0 1 9 9h-3a6 6 0 0 0-6-6V3Z" />
          </svg>
        ) : (
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="h-5 w-5" aria-hidden="true">
            <path d="M9 7H4v5" />
            <path d="M4 12a8 8 0 1 0 2.34-5.66L4 9" />
          </svg>
        )}
        <span className="pointer-events-none absolute -bottom-8 left-1/2 z-10 w-max -translate-x-1/2 rounded-md bg-[#0F172A] px-2 py-1 text-xs text-[#94A3B8] opacity-0 shadow-md transition-opacity group-hover:opacity-100">
          {labels.undo}
        </span>
      </button>
    </div>
  );
}

export default OrganizerActions;
