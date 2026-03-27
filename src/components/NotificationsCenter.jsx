function getItemTone(itemType) {
  if (itemType === 'organize') {
    return 'border-[#22C55E]/45 bg-[#22C55E]/12 text-[#b8ffcf]';
  }

  if (itemType === 'restore') {
    return 'border-rose-500/45 bg-rose-500/12 text-rose-200';
  }

  if (itemType === 'error') {
    return 'border-rose-500/45 bg-rose-500/12 text-rose-200';
  }

  return 'border-[#3B82F6]/45 bg-[#3B82F6]/12 text-[#bfdbfe]';
}

function NotificationsCenter({ labels, notifications, isOpen, onToggle, onClose, onClear }) {
  return (
    <>
      <button
        type="button"
        onClick={onToggle}
        aria-label={labels.notificationsTitle}
        className="fixed right-5 top-5 z-50 inline-flex h-11 w-11 items-center justify-center rounded-full border border-white/10 bg-[#1E293B]/95 text-[#F8FAFC] shadow-md transition-colors hover:bg-[#2b3a4f]"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="h-5 w-5" aria-hidden="true">
          <path d="M15 17h5l-1.4-1.4A2 2 0 0 1 18 14.2V10a6 6 0 1 0-12 0v4.2a2 2 0 0 1-.6 1.4L4 17h5" />
          <path d="M10 17a2 2 0 0 0 4 0" />
        </svg>
        {notifications.length > 0 && (
          <span className="absolute -right-1 -top-1 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-[#EF4444] px-1 text-[10px] font-bold text-white">
            {notifications.length > 99 ? '99+' : notifications.length}
          </span>
        )}
      </button>

      <div
        onClick={onClose}
        className={`fixed inset-0 z-40 bg-black/40 transition-opacity duration-300 ${
          isOpen ? 'opacity-100' : 'pointer-events-none opacity-0'
        }`}
      />

      <aside
        className={`fixed right-0 top-0 z-50 h-full w-[min(92vw,380px)] border-l border-white/10 bg-[#111827] p-5 shadow-2xl transition-transform duration-300 ${
          isOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[#F8FAFC]">{labels.notificationsTitle}</h2>
          <button
            type="button"
            onClick={onClear}
            className="text-sm font-semibold text-[#94A3B8] transition-colors hover:text-[#F8FAFC]"
          >
            {labels.notificationsClear}
          </button>
        </div>

        <div className="max-h-[calc(100%-56px)] space-y-2 overflow-y-auto pr-1">
          {notifications.length === 0 ? (
            <p className="text-sm text-[#94A3B8]">{labels.notificationsEmpty}</p>
          ) : (
            notifications.map((item) => (
              <div key={item.id} className={`rounded-xl border px-3 py-3 text-sm ${getItemTone(item.type)}`}>
                <p className="leading-relaxed">{item.message}</p>
                <p className="mt-1 text-[11px] opacity-80">{item.time}</p>
              </div>
            ))
          )}
        </div>
      </aside>
    </>
  );
}

export default NotificationsCenter;
