function FolderPathsPanel({
  sourceLabel,
  destinationLabel,
  sourcePath,
  destinationPath,
  hasDestination,
  destinationSelectHintPrefix,
  destinationSelectHintAction,
  onSelectDestinationFolder
}) {
  const abbreviatedDestinationPath =
    destinationPath.length > 72
      ? `${destinationPath.slice(0, 32)}...${destinationPath.slice(-32)}`
      : destinationPath;

  return (
    <div className="grid gap-4 md:grid-cols-2">
      <div className="flex h-36 flex-col rounded-2xl border border-white/10 bg-[#0F172A]/65 px-4 py-4 shadow-sm">
        <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-[#94A3B8]">{sourceLabel}</p>
        <p className="min-h-12 overflow-hidden text-ellipsis whitespace-nowrap text-sm text-[#F8FAFC]" title={sourcePath}>
          {sourcePath}
        </p>
      </div>

      <div className="flex h-36 flex-col rounded-2xl border border-white/10 bg-[#0F172A]/65 px-4 py-4 shadow-sm">
        <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-[#94A3B8]">{destinationLabel}</p>
        <p
          title={destinationPath}
          className="min-h-12 overflow-hidden text-ellipsis whitespace-nowrap text-sm text-[#F8FAFC]"
        >
          {abbreviatedDestinationPath}
        </p>

        <p className="mt-2 text-sm text-[#94A3B8]">
          {destinationSelectHintPrefix ? `${destinationSelectHintPrefix} ` : ''}
          <button
            type="button"
            onClick={onSelectDestinationFolder}
            className="font-semibold text-[#3B82F6] underline underline-offset-2 transition-colors hover:text-[#60a5fa]"
          >
            {destinationSelectHintAction}
          </button>
        </p>
      </div>
    </div>
  );
}

export default FolderPathsPanel;
