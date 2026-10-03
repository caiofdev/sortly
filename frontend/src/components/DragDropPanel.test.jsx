// função       | CC | casos
// handleDrop   |  2 | com File.path (Electron); sem File.path (Wails: caminho vem pelo gateway)
// dragover/leave |  1 | destaque liga e desliga
//
// O painel é a área de drop do Wails: estilo --wails-drop-target: drop.

import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import DragDropPanel from './DragDropPanel';

const labels = organizerCopy['pt-BR'];

function renderPanel() {
  const onResolveDroppedPath = vi.fn();
  render(
    <DragDropPanel
      isLoading={false}
      labels={labels}
      onResolveDroppedPath={onResolveDroppedPath}
      onSelectSourceFolder={vi.fn()}
    />
  );
  const panel = screen.getByText(labels.dropTitle).parentElement;
  return { panel, onResolveDroppedPath };
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

  it('versão Electron: usa File.path do primeiro arquivo', () => {
    const { panel, onResolveDroppedPath } = renderPanel();
    fireEvent.drop(panel, {
      dataTransfer: { files: [{ path: 'C:/a.txt' }, { path: 'C:/b.txt' }] }
    });
    expect(onResolveDroppedPath).toHaveBeenCalledWith('C:/a.txt');
  });

  it('versão Wails: sem File.path, não faz nada (o caminho vem pelo gateway)', () => {
    const { panel, onResolveDroppedPath } = renderPanel();
    fireEvent.drop(panel, { dataTransfer: { files: [{ name: 'a.txt' }] } });
    expect(onResolveDroppedPath).not.toHaveBeenCalled();
    expect(panel.className).toContain('border-white/20');
  });
});
