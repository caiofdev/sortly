import { useEffect, useRef, useState } from 'react';
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
  // Com o mouse ou o foco no toast, o prazo para: dá tempo de ler um caminho
  // longo ou chegar ao botão de fechar (WCAG 2.2.1, #75).
  const [paused, setPaused] = useState(false);

  // O pai recria onClose a cada estado novo; pela ref, o prazo de 4 s não
  // recomeça a cada atualização da tela (#75).
  const close = useRef(onClose);
  useEffect(() => {
    close.current = onClose;
  });

  useEffect(() => {
    if (sticky || paused) return undefined;
    const timer = setTimeout(() => close.current(), TOAST_MS);
    return () => clearTimeout(timer);
  }, [sticky, paused]);

  return (
    <div
      className={variant.className}
      role={variant.role}
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
      onFocus={() => setPaused(true)}
      onBlur={() => setPaused(false)}
    >
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
