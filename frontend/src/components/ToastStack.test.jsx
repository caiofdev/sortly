import { describe, expect, it } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import ToastStack, { MAX_TOASTS } from './ToastStack';

// Como no backend: a mais recente no topo, com id maior.
const notice = (id, kind = 'error', title = `aviso ${id}`) => ({ id, kind, title, text: '' });
const newestFirst = (...ids) => ids.sort((a, b) => b - a).map((id) => notice(id));

const titles = () => screen.queryAllByText(/^aviso /).map((el) => el.textContent);

function renderStack(notifications) {
  const view = render(<ToastStack notifications={notifications} closeLabel="Fechar aviso" />);
  const update = (next) =>
    view.rerender(<ToastStack notifications={next} closeLabel="Fechar aviso" />);
  return { ...view, update };
}

describe('ToastStack', () => {
  it('sem notificações: nenhum toast, mas a região anunciável existe', () => {
    const { container } = renderStack([]);
    expect(titles()).toEqual([]);
    expect(container.firstChild).toHaveAttribute('aria-live', 'polite');
  });

  it('cada notificação nova vira toast, a mais recente primeiro', () => {
    const { update } = renderStack(newestFirst(1));
    update(newestFirst(1, 2));
    expect(titles()).toEqual(['aviso 2', 'aviso 1']);
  });

  it('uma notificação já vista não volta depois de fechada', () => {
    const { update } = renderStack(newestFirst(1));
    fireEvent.click(screen.getByRole('button', { name: 'Fechar aviso' }));
    update(newestFirst(1));
    expect(titles()).toEqual([]);
  });

  it.each([
    [MAX_TOASTS - 1, MAX_TOASTS - 1],
    [MAX_TOASTS, MAX_TOASTS],
    [MAX_TOASTS + 1, MAX_TOASTS]
  ])('%i notificações novas: %i toasts, as mais recentes', (count, shown) => {
    const ids = Array.from({ length: count }, (_, i) => i + 1);
    renderStack(newestFirst(...ids));
    expect(titles()).toHaveLength(shown);
    expect(titles()[0]).toBe(`aviso ${count}`);
  });

  it('o texto segue o item atual (troca de idioma)', () => {
    const { update } = renderStack([notice(1, 'error', 'aviso pt')]);
    update([notice(1, 'error', 'aviso en')]);
    expect(screen.getByText('aviso en')).toBeInTheDocument();
  });

  it('limpar as notificações tira os toasts delas', () => {
    const { update } = renderStack(newestFirst(1));
    update([]);
    expect(titles()).toEqual([]);
  });
});
