import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import DragDropPanel from './DragDropPanel';

const labels = organizerCopy['pt-BR'];
const ACTIVE = 'st-drop--active';

function renderPanel({ isLoading = false } = {}) {
  const onSelectSourceFolder = vi.fn();
  const { container } = render(
    <DragDropPanel
      isLoading={isLoading}
      labels={labels}
      onSelectSourceFolder={onSelectSourceFolder}
    />
  );
  return { panel: container.querySelector('.st-drop'), onSelectSourceFolder };
}

describe('DragDropPanel', () => {
  it('é a área de drop do Wails', () => {
    const { panel } = renderPanel();
    expect(panel.style.getPropertyValue('--wails-drop-target')).toBe('drop');
  });

  it('arrastando por cima: destaque e título "solte"; ao sair, volta ao repouso', () => {
    const { panel } = renderPanel();
    fireEvent.dragOver(panel);
    expect(panel).toHaveClass(ACTIVE);
    expect(screen.getByText(labels.dropActive)).toBeInTheDocument();
    fireEvent.dragLeave(panel);
    expect(panel).not.toHaveClass(ACTIVE);
    expect(screen.getByText(labels.dropTitle)).toBeInTheDocument();
  });

  // Regressão (#60): entrar num elemento interno dispara dragleave no painel.
  it('arrastar sobre um elemento interno não apaga o destaque', () => {
    const { panel } = renderPanel();
    const inner = screen.getByRole('button', { name: labels.dropSelectHintAction });
    fireEvent.dragEnter(panel);
    fireEvent.dragEnter(inner);
    fireEvent.dragLeave(panel);
    expect(panel).toHaveClass(ACTIVE);
    fireEvent.dragLeave(inner);
    expect(panel).not.toHaveClass(ACTIVE);
  });

  it('dragleave sem dragenter não deixa o contador negativo', () => {
    const { panel } = renderPanel();
    fireEvent.dragLeave(panel);
    fireEvent.dragEnter(panel);
    expect(panel).toHaveClass(ACTIVE);
    fireEvent.dragLeave(panel);
    expect(panel).not.toHaveClass(ACTIVE);
  });

  it('soltar desliga o destaque', () => {
    const { panel } = renderPanel();
    fireEvent.dragOver(panel);
    fireEvent.drop(panel, { dataTransfer: { files: [{ name: 'a.txt' }] } });
    expect(panel).not.toHaveClass(ACTIVE);
  });

  it.each([
    [false, false],
    [true, true]
  ])('carregando = %s: link desabilitado = %s', (isLoading, disabled) => {
    const { panel, onSelectSourceFolder } = renderPanel({ isLoading });
    const link = screen.getByRole('button', { name: labels.dropSelectHintAction });
    expect(link.disabled).toBe(disabled);
    expect(panel).toHaveAttribute('aria-busy', String(isLoading));
    fireEvent.click(link);
    expect(onSelectSourceFolder).toHaveBeenCalledTimes(disabled ? 0 : 1);
  });
});
