import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import FolderPathsPanel from './FolderPathsPanel';

function renderPanel(props = {}) {
  const onSelectDestinationFolder = vi.fn();
  render(
    <FolderPathsPanel
      sourceLabel="Origem"
      destinationLabel="Destino"
      sourcePath="C:\origem"
      destinationPath="C:\destino"
      destinationSelectHintPrefix="Ou"
      destinationSelectHintAction="Selecione"
      onSelectDestinationFolder={onSelectDestinationFolder}
      {...props}
    />
  );
  return { onSelectDestinationFolder };
}

describe('FolderPathsPanel', () => {
  it.each([
    [72, false],
    [73, true]
  ])(
    'destino com %i caracteres: abreviado = %s, title com o caminho completo',
    (length, abbreviated) => {
      const path = `C:\\${'x'.repeat(length - 3)}`;
      renderPanel({ destinationPath: path });
      expect(screen.getByTitle(path).textContent.includes('...')).toBe(abbreviated);
    }
  );

  it.each([
    ['Ou', 'Ou Selecione'],
    ['', 'Selecione']
  ])('prefixo "%s" da dica do destino', (prefix, text) => {
    renderPanel({ destinationSelectHintPrefix: prefix });
    expect(screen.getByRole('button', { name: 'Selecione' }).parentElement.textContent).toBe(text);
  });

  it('o link da dica escolhe o destino', () => {
    const { onSelectDestinationFolder } = renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'Selecione' }));
    expect(onSelectDestinationFolder).toHaveBeenCalled();
  });
});
