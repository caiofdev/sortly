    function OrganizerActions({
  labels,
  isLoading,
  hasUndo,
  hasSource,
  onSelectDestinationFolder,
  onOrganizeFiles,
  onUndoLastOrganization
}) {
  return (
    <>
      <button
        type="button"
        onClick={onSelectDestinationFolder}
        className="rounded-2xl bg-[#334155] px-6 py-4 text-base font-semibold text-[#F8FAFC] shadow-sm transition-all hover:-translate-y-0.5 hover:bg-[#3b4a5d] hover:shadow-md"
      >
        {labels.destinationButton}
      </button>

      <button
        type="button"
        onClick={onOrganizeFiles}
        disabled={isLoading || !hasSource}
        className="rounded-2xl bg-[#3B82F6] px-6 py-4 text-lg font-semibold text-white shadow-[0_10px_30px_rgba(59,130,246,0.35)] transition-all hover:-translate-y-0.5 hover:bg-[#4f8ff7] disabled:cursor-not-allowed disabled:bg-[#2b3b57] disabled:text-slate-300 disabled:shadow-none"
      >
        {isLoading ? labels.organizing : labels.organize}
      </button>

      <button
        type="button"
        onClick={onUndoLastOrganization}
        disabled={isLoading || !hasUndo}
        className="rounded-2xl border border-[#3B82F6]/45 bg-transparent px-6 py-4 text-base font-semibold text-[#F8FAFC] transition-all hover:-translate-y-0.5 hover:border-[#3B82F6] hover:bg-[#3B82F6]/12 disabled:cursor-not-allowed disabled:border-slate-600/60 disabled:text-slate-400"
      >
        {labels.undo}
      </button>
    </>
  );
}

export default OrganizerActions;
