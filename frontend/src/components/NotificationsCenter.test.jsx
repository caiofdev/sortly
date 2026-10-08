import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import NotificationsCenter, { dotClass } from './NotificationsCenter';

const labels = organizerCopy['pt-BR'];
const items = (types) =>
  types.map((type, i) => ({ id: i, type, message: `aviso ${i}`, time: '09:07' }));

function renderCenter(props = {}) {
  const handlers = { onToggle: vi.fn(), onClear: vi.fn() };
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

describe('dotClass', () => {
  it.each([
    ['organize', 'st-notif__dot st-notif__dot--ok'],
    ['error', 'st-notif__dot st-notif__dot--err'],
    ['restore', 'st-notif__dot'],
    ['info', 'st-notif__dot']
  ])('%s → %s', (type, expected) => {
    expect(dotClass(type)).toBe(expected);
  });
});

describe('NotificationsCenter', () => {
  it.each([
    [false, 'false', 0],
    [true, 'true', 1]
  ])('aberto = %s: sino com aria-expanded %s e %i painel', (isOpen, expanded, panels) => {
    renderCenter({ isOpen });
    expect(screen.getByRole('button', { name: labels.notificationsTitle })).toHaveAttribute(
      'aria-expanded',
      expanded
    );
    expect(screen.queryAllByRole('dialog')).toHaveLength(panels);
  });

  it('sem notificações: texto de vazio', () => {
    renderCenter({ isOpen: true });
    expect(screen.getByText(labels.notificationsEmpty)).toBeInTheDocument();
  });

  it('lista cada aviso com o ponto do seu status e a hora', () => {
    const { container } = renderCenter({
      isOpen: true,
      notifications: items(['organize', 'error', 'info'])
    });
    expect(screen.queryByText(labels.notificationsEmpty)).not.toBeInTheDocument();
    expect(screen.getAllByText(/^aviso /)).toHaveLength(3);
    expect(container.querySelectorAll('.st-notif__dot--ok')).toHaveLength(1);
    expect(container.querySelectorAll('.st-notif__dot--err')).toHaveLength(1);
    expect(screen.getAllByText('09:07')).toHaveLength(3);
  });

  it('o sino alterna e "Limpar" limpa', () => {
    const { onToggle, onClear } = renderCenter({ isOpen: true });
    fireEvent.click(screen.getByRole('button', { name: labels.notificationsTitle }));
    fireEvent.click(screen.getByRole('button', { name: labels.notificationsClear }));
    expect(onToggle).toHaveBeenCalled();
    expect(onClear).toHaveBeenCalled();
  });
});
