// função       | CC | casos
// handleDrop   |  1 | desliga o destaque (o caminho vem pelo gateway)
// dragover/leave |  1 | destaque liga e desliga
//
// O painel é a área de drop do Wails: estilo --wails-drop-target: drop.

import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import DragDropPanel from './DragDropPanel';

const labels = organizerCopy['pt-BR'];

function renderPanel() {
  render(<DragDropPanel isLoading={false} labels={labels} onSelectSourceFolder={vi.fn()} />);
  return { panel: screen.getByText(labels.dropTitle).parentElement };
}

describe('DragDropPanel', () => {
  it('é a área de drop do Wails', () => {
    const { panel } = renderPanel();
    expect(panel.style.getPropertyValue('--wails-drop-target')).toBe('drop');
  });

  it('destaca ao arrastar por cima e volta ao sair', () => {
    const { panel } = renderPanel();
    fireEvent.dragOver(panel);
    expect(panel.className).toContain('border-[#3B82F6]');
    fireEvent.dragLeave(panel);
    expect(panel.className).toContain('border-white/20');
  });

  it('soltar desliga o destaque', () => {
    const { panel } = renderPanel();
    fireEvent.dragOver(panel);
    fireEvent.drop(panel, { dataTransfer: { files: [{ name: 'a.txt' }] } });
    expect(panel.className).toContain('border-white/20');
  });
});
