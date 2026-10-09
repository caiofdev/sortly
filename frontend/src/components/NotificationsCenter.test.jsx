import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import NotificationsCenter, { dotClass } from './NotificationsCenter';

const labels = organizerCopy['pt-BR'];
const items = (kinds) =>
  kinds.map((kind, i) => ({ id: i, kind, title: `aviso ${i}`, text: '', time: '09:07' }));

function renderCenter(props = {}) {
  const handlers = { onToggle: vi.fn(), onClose: vi.fn(), onMarkRead: vi.fn(), onClear: vi.fn() };
  const view = render(
    <NotificationsCenter
      labels={labels}
      notifications={[]}
      unread={false}
      isOpen={false}
      {...handlers}
      {...props}
    />
  );
  return { ...view, ...handlers };
}

const bell = () => screen.getByRole('button', { name: labels.notificationsTitle });

describe('dotClass', () => {
  it.each([
    ['success', 'st-notif__dot st-notif__dot--ok'],
    ['error', 'st-notif__dot st-notif__dot--err'],
    ['info', 'st-notif__dot'],
    ['desconhecido', 'st-notif__dot']
  ])('%s → %s', (kind, expected) => {
    expect(dotClass(kind)).toBe(expected);
  });
});

describe('NotificationsCenter', () => {
  it.each([
    [false, 'false', 0],
    [true, 'true', 1]
  ])('aberto = %s: sino com aria-expanded %s e %i painel', (isOpen, expanded, panels) => {
    renderCenter({ isOpen });
    expect(bell()).toHaveAttribute('aria-expanded', expanded);
    expect(screen.queryAllByRole('dialog')).toHaveLength(panels);
  });

  it.each([
    [false, 0, null],
    [true, 1, labels.notificationsUnread]
  ])('não lidas = %s: %i ponto amarelo e descrição %s', (unread, dots, description) => {
    const { container } = renderCenter({ unread });
    expect(container.querySelectorAll('.st-icon-btn__dot')).toHaveLength(dots);
    if (description) {
      expect(bell()).toHaveAccessibleDescription(description);
    } else {
      expect(bell()).not.toHaveAttribute('aria-describedby');
    }
  });

  it('sem notificações: texto de vazio', () => {
    renderCenter({ isOpen: true });
    expect(screen.getByText(labels.notificationsEmpty)).toBeInTheDocument();
  });

  it('lista cada aviso com o ponto do seu status, o detalhe e a hora', () => {
    const list = items(['success', 'error', 'info']);
    list[0].text = 'Na pasta C:\\destino.';
    const { container } = renderCenter({ isOpen: true, notifications: list });
    expect(screen.queryByText(labels.notificationsEmpty)).not.toBeInTheDocument();
    expect(screen.getAllByText(/^aviso /)).toHaveLength(3);
    expect(container.querySelectorAll('.st-notif__text')).toHaveLength(1);
    expect(screen.getByText('Na pasta C:\\destino.')).toBeInTheDocument();
    expect(container.querySelectorAll('.st-notif__dot--ok')).toHaveLength(1);
    expect(container.querySelectorAll('.st-notif__dot--err')).toHaveLength(1);
    expect(screen.getAllByText('09:07')).toHaveLength(3);
  });

  it('o sino alterna; "Marcar como lidas" e "Limpar" chamam as ações', () => {
    const { onToggle, onMarkRead, onClear } = renderCenter({ isOpen: true });
    fireEvent.click(bell());
    fireEvent.click(screen.getByRole('button', { name: labels.notificationsMarkRead }));
    fireEvent.click(screen.getByRole('button', { name: labels.notificationsClear }));
    expect(onToggle).toHaveBeenCalled();
    expect(onMarkRead).toHaveBeenCalled();
    expect(onClear).toHaveBeenCalled();
  });

  it('ao abrir, o foco vai para o painel', () => {
    renderCenter({ isOpen: true });
    expect(screen.getByRole('dialog')).toHaveFocus();
  });

  it.each([
    ['Escape', 1],
    ['Enter', 0]
  ])('tecla %s no painel: fecha %i vez', (key, closes) => {
    const { onClose } = renderCenter({ isOpen: true });
    fireEvent.keyDown(screen.getByRole('dialog'), { key });
    expect(onClose).toHaveBeenCalledTimes(closes);
    if (closes) expect(bell()).toHaveFocus();
  });
});
