import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import PathField from './PathField';

const LONG = `C:\\${'pasta-com-nome-longo\\'.repeat(5)}arquivos`;

describe('PathField', () => {
  it.each([
    ['C:\\origem', 'C:\\origem', false],
    ['', 'Nenhuma pasta selecionada', true]
  ])('caminho "%s" mostra "%s" (vazio = %s)', (path, shown, empty) => {
    render(<PathField label="Origem" path={path} emptyText="Nenhuma pasta selecionada" />);
    expect(screen.getByText(shown).classList.contains('st-path__value--empty')).toBe(empty);
  });

  it('caminho longo aparece inteiro, sem "...", e com o title completo', () => {
    render(<PathField label="Destino" path={LONG} emptyText="" />);
    expect(screen.getByTitle(LONG).textContent).toBe(LONG);
  });

  it.each([
    [undefined, 0],
    ['Trocar pasta de destino', 1]
  ])('ação "%s": %i link(s)', (actionLabel, links) => {
    const onAction = vi.fn();
    render(
      <PathField
        label="Destino"
        path=""
        emptyText="vazio"
        actionLabel={actionLabel}
        onAction={onAction}
      />
    );
    const buttons = screen.queryAllByRole('button');
    expect(buttons).toHaveLength(links);
    buttons.forEach((button) => fireEvent.click(button));
    expect(onAction).toHaveBeenCalledTimes(links);
  });

  it('ação desabilitada não dispara', () => {
    const onAction = vi.fn();
    render(
      <PathField
        label="Destino"
        path=""
        emptyText="vazio"
        actionLabel="Trocar"
        onAction={onAction}
        disabled
      />
    );
    fireEvent.click(screen.getByRole('button', { name: 'Trocar' }));
    expect(onAction).not.toHaveBeenCalled();
  });
});
