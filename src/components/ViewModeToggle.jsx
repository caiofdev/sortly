function ViewModeToggle({ viewMode, onChange, labels }) {
  const options = [
    { value: 'default', label: labels.defaultMode },
    { value: 'drag-drop', label: labels.dragDropMode }
  ];

  return (
    <div className="inline-flex rounded-2xl border border-white/10 bg-[#0F172A]/60 p-1">
      {options.map((option) => {
        const isActive = viewMode === option.value;
        return (
          <button
            key={option.value}
            type="button"
            onClick={() => onChange(option.value)}
            className={`rounded-xl px-3 py-2 text-sm font-semibold transition-colors ${
              isActive ? 'bg-[#3B82F6] text-white' : 'text-[#94A3B8] hover:bg-white/5 hover:text-[#F8FAFC]'
            }`}
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}

export default ViewModeToggle;
