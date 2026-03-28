function OrganizerSettingsPanel({ isOpen, onToggle, labels, options, onOptionChange }) {
  const selectedCount = Object.values(options).filter(Boolean).length;

  const settingsOptions = [
    { key: 'byDuration', label: labels.settingsByDuration },
    { key: 'byPages', label: labels.settingsByPages },
    { key: 'byResolution', label: labels.settingsByResolution },
    { key: 'byDate', label: labels.settingsByDate },
    { key: 'bySize', label: labels.settingsBySize },
    { key: 'byExtension', label: labels.settingsByExtension }
  ];

  return (
    <div className="relative">
      <button
        type="button"
        onClick={onToggle}
        aria-label={labels.settingsTitle}
        className="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-[#0F172A]/60 text-[#F8FAFC] transition-colors hover:bg-[#1f2c3f]"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="h-5 w-5" aria-hidden="true">
          <path d="M12 15.5A3.5 3.5 0 1 0 12 8.5a3.5 3.5 0 0 0 0 7Z" />
          <path d="M19.4 15a1 1 0 0 0 .2 1.1l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1 1 0 0 0-1.1-.2 1 1 0 0 0-.6.9V20a2 2 0 1 1-4 0v-.2a1 1 0 0 0-.6-.9 1 1 0 0 0-1.1.2l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1 1 0 0 0 .2-1.1 1 1 0 0 0-.9-.6H4a2 2 0 1 1 0-4h.2a1 1 0 0 0 .9-.6 1 1 0 0 0-.2-1.1l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1 1 0 0 0 1.1.2H9a1 1 0 0 0 .6-.9V4a2 2 0 1 1 4 0v.2a1 1 0 0 0 .6.9h.2a1 1 0 0 0 1.1-.2l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1 1 0 0 0-.2 1.1v.2a1 1 0 0 0 .9.6H20a2 2 0 1 1 0 4h-.2a1 1 0 0 0-.9.6V15Z" />
        </svg>
      </button>

      <div
        className={`absolute right-0 top-12 z-30 w-[min(92vw,330px)] rounded-2xl border border-white/10 bg-[#111827] p-4 shadow-xl transition-all duration-200 ${
          isOpen ? 'translate-y-0 opacity-100' : 'pointer-events-none -translate-y-2 opacity-0'
        }`}
      >
        <h3 className="text-sm font-semibold text-[#F8FAFC]">{labels.settingsTitle}</h3>
        <p className="mt-1 text-xs text-[#94A3B8]">{labels.settingsSubtitle}</p>

        <div className="mt-3 space-y-2">
          {settingsOptions.map((item) => {
            const isChecked = Boolean(options[item.key]);
            const shouldDisable = isChecked && selectedCount === 1;

            return (
            <label
              key={item.key}
              className={`flex items-center justify-between rounded-lg bg-white/5 px-3 py-2 text-sm ${
                shouldDisable ? 'cursor-not-allowed opacity-70' : 'cursor-pointer'
              }`}
            >
              <span className="text-[#F8FAFC]">{item.label}</span>
              <input
                type="checkbox"
                checked={isChecked}
                disabled={shouldDisable}
                onChange={(event) => onOptionChange(item.key, event.target.checked)}
                className="h-4 w-4 accent-[#3B82F6]"
              />
            </label>
            );
          })}
        </div>
      </div>
    </div>
  );
}

export default OrganizerSettingsPanel;
