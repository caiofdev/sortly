import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import NotificationsCenter, { getItemTone } from './NotificationsCenter';

const labels = organizerCopy['pt-BR'];
const items = (n) =>
  Array.from({ length: n }, (_, i) => ({
    id: i,
    type: 'info',
    message: `aviso ${i}`,
    time: '09:07'
  }));

function renderCenter(props = {}) {
  const handlers = { onToggle: vi.fn(), onClose: vi.fn(), onClear: vi.fn() };
  const view = render(
    <NotificationsCenter
      labels={labels}
      notifications={[]}
      isOpen={false}
      {...handlers}
      {...props}
    />
  );
  return { ...view, ...handlers };
}

describe('getItemTone', () => {
  it.each([
    ['organize', '#22C55E'],
    ['restore', 'rose-500'],
    ['error', 'rose-500'],
    ['info', '#3B82F6'],
    ['qualquer', '#3B82F6']
  ])('%s → tom com %s', (type, color) => {
    expect(getItemTone(type)).toContain(color);
  });
});

describe('NotificationsCenter', () => {
  it('sem notificações: texto de vazio e sem contador', () => {
    renderCenter();
    expect(screen.getByText(labels.notificationsEmpty)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: labels.notificationsTitle }).textContent).toBe('');
  });

  it.each([
    [1, '1'],
    [99, '99'],
    [100, '99+']
  ])('%i notificação(ões): contador "%s"', (n, badge) => {
    renderCenter({ notifications: items(n) });
    expect(screen.getByRole('button', { name: labels.notificationsTitle }).textContent).toBe(badge);
    expect(screen.getAllByText(/^aviso /)).toHaveLength(n);
  });

  it.each([
    [true, 'translate-x-0'],
    [false, 'translate-x-full']
  ])('aberto = %s: painel com %s', (isOpen, translate) => {
    const { container } = renderCenter({ isOpen });
    expect(container.querySelector('aside').className).toContain(translate);
  });

  it('sino alterna, fundo fecha e "limpar" limpa', () => {
    const { container, onToggle, onClose, onClear } = renderCenter({ isOpen: true });
    fireEvent.click(screen.getByRole('button', { name: labels.notificationsTitle }));
    fireEvent.click(container.querySelector('aside').previousElementSibling);
    fireEvent.click(screen.getByRole('button', { name: labels.notificationsClear }));
    expect(onToggle).toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
    expect(onClear).toHaveBeenCalled();
  });
});
