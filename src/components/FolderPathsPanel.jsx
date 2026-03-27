function FolderPathsPanel({ sourceLabel, destinationLabel, sourcePath, destinationPath }) {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      <div className="rounded-2xl border border-white/10 bg-[#0F172A]/65 px-4 py-4 shadow-sm">
        <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-[#94A3B8]">{sourceLabel}</p>
        <p className="min-h-12 break-all text-sm text-[#F8FAFC]">{sourcePath}</p>
      </div>

      <div className="rounded-2xl border border-white/10 bg-[#0F172A]/65 px-4 py-4 shadow-sm">
        <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-[#94A3B8]">{destinationLabel}</p>
        <p className="min-h-12 break-all text-sm text-[#F8FAFC]">{destinationPath}</p>
      </div>
    </div>
  );
}

export default FolderPathsPanel;
