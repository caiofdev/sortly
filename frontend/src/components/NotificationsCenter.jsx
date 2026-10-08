import { useEffect, useId, useRef } from 'react';
import { BellIcon } from './icons';

// Quadradinho de status do item: verde para sucesso, vermelho para erro e
// neutro para o resto (design system, Notifications; #75).
const DOT_BY_KIND = {
  success: 'st-notif__dot st-notif__dot--ok',
  error: 'st-notif__dot st-notif__dot--err'
};

export function dotClass(kind) {
  return DOT_BY_KIND[kind] ?? 'st-notif__dot';
}

function NotificationsCenter({
  labels,
  notifications,
  unread,
  isOpen,
  onToggle,
  onClose,
  onMarkRead,
  onClear
}) {
  const bell = useRef(null);
  const panel = useRef(null);
  const unreadId = useId();

  // Ao abrir, o foco entra no painel, para o leitor de tela anunciar o diálogo e o
  // Esc funcionar sem clicar nele antes (#75).
  useEffect(() => {
    if (isOpen) panel.current.focus();
  }, [isOpen]);

  const handleKeyDown = (event) => {
    if (event.key !== 'Escape') return;
    onClose();
    bell.current.focus();
  };

  return (
    <>
      <button
        ref={bell}
        type="button"
        className="st-icon-btn st-icon-btn--round"
        aria-label={labels.notificationsTitle}
        aria-expanded={isOpen}
        aria-describedby={unread ? unreadId : undefined}
        onClick={onToggle}
      >
        <BellIcon />
        {unread && (
          <span className="st-icon-btn__dot">
            <span id={unreadId} className="st-sr-only">
              {labels.notificationsUnread}
            </span>
          </span>
        )}
      </button>

      {isOpen && (
        <div
          ref={panel}
          className="st-notif"
          role="dialog"
          aria-label={labels.notificationsTitle}
          tabIndex={-1}
          onKeyDown={handleKeyDown}
        >
          <div className="st-notif__head">
            <p>{labels.notificationsTitle}</p>
            <div className="st-notif__actions">
              <button type="button" className="st-link" onClick={onMarkRead}>
                {labels.notificationsMarkRead}
              </button>
              <button type="button" className="st-link" onClick={onClear}>
                {labels.notificationsClear}
              </button>
            </div>
          </div>
          <div className="st-notif__list">
            {notifications.length === 0 && (
              <p className="st-notif__empty">{labels.notificationsEmpty}</p>
            )}
            {notifications.map((item) => (
              <div key={item.id} className="st-notif__item">
                <span className={dotClass(item.kind)} />
                <div>
                  <p className="st-notif__title">{item.title}</p>
                  {item.text && <p className="st-notif__text">{item.text}</p>}
                  <p className="st-notif__when">{item.time}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </>
  );
}

export default NotificationsCenter;
