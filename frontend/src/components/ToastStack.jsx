import { useEffect, useRef, useState } from 'react';
import Toast from './Toast';

// Um por vez, como no protótipo: empilhados, cobriam os botões Organizar e
// Desfazer na janela padrão, e o de erro só sai quando o usuário fecha (#75).
export const MAX_TOASTS = 1;

// Toda notificação com id maior que o último visto vira toast; quais estão na
// tela é estado de UI, e o texto vem do item atual para seguir o idioma (ADR 0005, #75).
function ToastStack({ notifications, closeLabel }) {
  const [ids, setIds] = useState([]);
  const lastSeen = useRef(0);

  useEffect(() => {
    const fresh = notifications.filter((n) => n.id > lastSeen.current).map((n) => n.id);
    if (fresh.length === 0) return;
    lastSeen.current = Math.max(...fresh);
    setIds((current) => [...fresh, ...current].slice(0, MAX_TOASTS));
  }, [notifications]);

  const dismiss = (id) => setIds((current) => current.filter((shown) => shown !== id));
  const visible = ids.map((id) => notifications.find((n) => n.id === id)).filter(Boolean);

  return (
    <div className="st-toasts" aria-live="polite">
      {visible.map((item) => (
        <Toast
          key={item.id}
          kind={item.kind}
          title={item.title}
          text={item.text}
          closeLabel={closeLabel}
          onClose={() => dismiss(item.id)}
        />
      ))}
    </div>
  );
}

export default ToastStack;
