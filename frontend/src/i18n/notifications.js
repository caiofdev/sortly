import feedbackCopy from './feedbackCopy';
import { getCopy, toLocale } from './language';

// Códigos que não são erros: cada um tem o próprio título e texto, com os dados
// da notificação (#75).
const notices = {
  RECOVERED_LAST_ORGANIZATION: (_n, copy) => copy.recovered,
  SOURCE_DROPPED: (n, copy) => copy.sourceDropped(n.path),
  SOURCE_REQUIRED: (_n, copy) => ({ title: copy.sourceRequired, text: '' }),
  ORGANIZE_DONE: (n, copy) => copy.organizeDone(n.organize),
  ORGANIZE_CANCELED: (n, copy) => copy.organizeCanceled(n.organize),
  UNDO_DONE: (n, copy) => copy.undoDone(n.undo)
};

const fallbackByAction = {
  selectSource: 'sourceSelectError',
  selectDestination: 'destinationSelectError',
  drop: 'droppedPathUnexpectedError',
  organize: 'organizeUnexpectedError',
  undo: 'undoUnexpectedError',
  settings: 'settingsSaveError',
  preview: 'previewError'
};

export function notificationText(notification, copy) {
  const notice = notices[notification.code];
  if (notice) {
    return notice(notification, copy);
  }
  const title =
    copy.errors[notification.code] ||
    copy[fallbackByAction[notification.action]] ||
    copy.unexpectedError;
  return { title, text: '' };
}

export function toNotificationItems(notifications, language) {
  const copy = getCopy(feedbackCopy, language);
  const locale = toLocale(language);
  return notifications.map((n) => ({
    id: n.id,
    kind: n.kind,
    ...notificationText(n, copy),
    time: new Date(n.at).toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit' })
  }));
}
