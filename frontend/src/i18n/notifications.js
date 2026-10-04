import feedbackCopy from './feedbackCopy';
import { getCopy, toLocale } from './language';

// Códigos que não são erros: cada um tem o próprio texto, com os dados da notificação.
const messages = {
  RECOVERED_LAST_ORGANIZATION: (_n, copy) => copy.recoveredLastOrganization,
  SOURCE_DROPPED: (n, copy) => `${copy.droppedPathSuccess} ${n.path}`,
  SOURCE_REQUIRED: (_n, copy) => copy.sourceRequired,
  ORGANIZE_DONE: (n, copy) => copy.organizeSuccess(n.organize),
  UNDO_DONE: (n, copy) => copy.undoSuccess(n.undo)
};

// Texto padrão de cada ação para erros sem tradução própria (ex.: UNEXPECTED).
const fallbackByAction = {
  selectSource: 'sourceSelectError',
  selectDestination: 'destinationSelectError',
  drop: 'droppedPathUnexpectedError',
  organize: 'organizeUnexpectedError',
  undo: 'undoUnexpectedError',
  settings: 'settingsSaveError'
};

export function notificationMessage(notification, copy) {
  const message = messages[notification.code];
  if (message) {
    return message(notification, copy);
  }
  return (
    copy.errors[notification.code] ||
    copy[fallbackByAction[notification.action]] ||
    copy.unexpectedError
  );
}

// Notificações do backend prontas para o NotificationsCenter, no idioma atual.
export function toNotificationItems(notifications, language) {
  const copy = getCopy(feedbackCopy, language);
  const locale = toLocale(language);
  return notifications.map((n) => ({
    id: n.id,
    type: n.kind,
    message: notificationMessage(n, copy),
    time: new Date(n.at).toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit' })
  }));
}
