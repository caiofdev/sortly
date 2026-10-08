import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, fireEvent, render, screen } from '@testing-library/react';
import Toast, { TOAST_MS } from './Toast';

function renderToast(props = {}) {
  const onClose = vi.fn();
  const view = render(
    <Toast
      kind="success"
      title="Pronto!"
      text=""
      closeLabel="Fechar aviso"
      onClose={onClose}
      {...props}
    />
  );
  return { ...view, onClose };
}

describe('Toast', () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it.each([
    ['success', 'st-toast--success', null],
    ['info', 'st-toast--info', null],
    ['error', 'st-toast--danger', 'alert'],
    ['desconhecido', 'st-toast--info', null]
  ])('%s: classe %s e role %s', (kind, className, role) => {
    const { container } = renderToast({ kind });
    const toast = container.firstChild;
    expect(toast).toHaveClass('st-toast', className);
    if (role) {
      expect(toast).toHaveAttribute('role', role);
    } else {
      expect(toast).not.toHaveAttribute('role');
    }
  });

  it.each([
    ['', 0],
    ['Na pasta C:\\destino.', 1]
  ])('texto "%s": %i parágrafo de detalhe', (text, count) => {
    const { container } = renderToast({ text });
    expect(container.querySelectorAll('.st-toast__text')).toHaveLength(count);
  });

  it('some sozinho em 4 s, nem um instante antes', () => {
    const { onClose } = renderToast();
    act(() => vi.advanceTimersByTime(TOAST_MS - 1));
    expect(onClose).not.toHaveBeenCalled();
    act(() => vi.advanceTimersByTime(1));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('erro fica até o usuário fechar', () => {
    const { onClose } = renderToast({ kind: 'error' });
    act(() => vi.advanceTimersByTime(TOAST_MS * 10));
    expect(onClose).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: 'Fechar aviso' }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('um onClose novo a cada render não recomeça o prazo', () => {
    const { rerender, onClose } = renderToast();
    act(() => vi.advanceTimersByTime(TOAST_MS - 1));
    const latest = vi.fn();
    rerender(<Toast kind="success" title="Pronto!" closeLabel="Fechar aviso" onClose={latest} />);
    act(() => vi.advanceTimersByTime(1));
    expect(onClose).not.toHaveBeenCalled();
    expect(latest).toHaveBeenCalledTimes(1);
  });
});
