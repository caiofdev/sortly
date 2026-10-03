import { useCallback, useRef, useState } from 'react';
import { toLocale } from '../i18n/language';
import { createId } from '../utils/createId';

export const MAX_NOTIFICATIONS = 80;

// Histórico de notificações, da mais recente para a mais antiga, limitado a
// MAX_NOTIFICATIONS. O horário é formatado no idioma do momento do aviso.
function useNotifications(language) {
  const [notifications, setNotifications] = useState([]);
  const languageRef = useRef(language);
  languageRef.current = language;

  const notify = useCallback((type, message) => {
    const time = new Date().toLocaleTimeString(toLocale(languageRef.current), {
      hour: '2-digit',
      minute: '2-digit'
    });
    const item = { id: createId(), type, message, time };
    setNotifications((previous) => [item, ...previous].slice(0, MAX_NOTIFICATIONS));
  }, []);

  const clearNotifications = useCallback(() => setNotifications([]), []);

  return { notifications, notify, clearNotifications };
}

export default useNotifications;
