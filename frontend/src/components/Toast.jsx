import { useEffect, useRef } from 'react';
import { AlertIcon, CheckIcon, CloseIcon, InfoIcon } from './icons';

export const TOAST_MS = 4000;

// O status vem do backend; erro usa role="alert" e fica até o usuário fechar
// (design system, Toast; #75).
const VARIANTS = {
  success: { className: 'st-toast st-toast--success', Icon: CheckIcon },
  info: { className: 'st-toast st-toast--info', Icon: InfoIcon },
  error: { className: 'st-toast st-toast--danger', Icon: AlertIcon, role: 'alert' }
};

function Toast({ kind, title, text, closeLabel, onClose }) {
  const variant = VARIANTS[kind] ?? VARIANTS.info;
  const sticky = Boolean(variant.role);

  // O pai recria onClose a cada estado novo; pela ref, o prazo de 4 s não
  // recomeça a cada atualização da tela (#75).
  const close = useRef(onClose);
  close.current = onClose;

  useEffect(() => {
    if (sticky) return undefined;
    const timer = setTimeout(() => close.current(), TOAST_MS);
    return () => clearTimeout(timer);
  }, [sticky]);

  return (
    <div className={variant.className} role={variant.role}>
      <span className="st-toast__icon">
        <variant.Icon />
      </span>
      <div className="st-toast__body">
        <p className="st-toast__title">{title}</p>
        {text && <p className="st-toast__text">{text}</p>}
      </div>
      <button type="button" className="st-toast__close" aria-label={closeLabel} onClick={onClose}>
        <CloseIcon />
      </button>
    </div>
  );
}

export default Toast;
