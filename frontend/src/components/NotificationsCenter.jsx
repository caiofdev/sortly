import { BellIcon } from './icons';

// Quadradinho de status do item: verde para organizar, vermelho para erro e
// neutro para o resto (design system, Notifications; #74).
const DOT_BY_TYPE = {
  organize: 'st-notif__dot st-notif__dot--ok',
  error: 'st-notif__dot st-notif__dot--err'
};

export function dotClass(itemType) {
  return DOT_BY_TYPE[itemType] ?? 'st-notif__dot';
}

function NotificationsCenter({ labels, notifications, isOpen, onToggle, onClear }) {
  return (
    <>
      <button
        type="button"
        className="st-icon-btn st-icon-btn--round"
        aria-label={labels.notificationsTitle}
        aria-expanded={isOpen}
        onClick={onToggle}
      >
        <BellIcon />
      </button>

      {isOpen && (
        <div className="st-notif" role="dialog" aria-label={labels.notificationsTitle}>
          <div className="st-notif__head">
            <p>{labels.notificationsTitle}</p>
            <button type="button" className="st-link" onClick={onClear}>
              {labels.notificationsClear}
            </button>
          </div>
          <div className="st-notif__list">
            {notifications.length === 0 && (
              <p className="st-notif__empty">{labels.notificationsEmpty}</p>
            )}
            {notifications.map((item) => (
              <div key={item.id} className="st-notif__item">
                <span className={dotClass(item.type)} />
                <div>
                  <p className="st-notif__title">{item.message}</p>
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
