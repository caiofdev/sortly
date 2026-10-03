// função                  | CC | casos
// getItemTone             |  2 | organize; restore e error (mesmo tom); info / desconhecido (padrão)
// OrganizerSettingsPanel  |  1 | 6 checkboxes na ordem; último marcado desabilitado
// FolderPathsPanel        |  1 | caminho de 73 caracteres abreviado; title com o caminho completo

import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { DEFAULT_OPTIONS } from '../domain/organizationOptions';
import organizerCopy from '../i18n/organizerCopy';
import FolderPathsPanel from './FolderPathsPanel';
import { getItemTone } from './NotificationsCenter';
import OrganizerSettingsPanel from './OrganizerSettingsPanel';

const labels = organizerCopy['pt-BR'];

describe('getItemTone', () => {
  it('organize é verde; restore e error compartilham o tom de alerta; o resto é azul', () => {
    expect(getItemTone('organize')).toContain('#22C55E');
    expect(getItemTone('restore')).toBe(getItemTone('error'));
    expect(getItemTone('info')).toContain('#3B82F6');
    expect(getItemTone('qualquer')).toBe(getItemTone('info'));
  });
});

describe('OrganizerSettingsPanel', () => {
  it('mostra os 6 critérios na ordem e desabilita o último marcado', () => {
    const onOptionChange = vi.fn();
    render(
      <OrganizerSettingsPanel
        isOpen
        labels={labels}
        options={DEFAULT_OPTIONS}
        onToggle={vi.fn()}
        onOptionChange={onOptionChange}
      />
    );
    const boxes = screen.getAllByRole('checkbox');
    expect(boxes.map((b) => b.closest('label').textContent)).toEqual([
      'Duração (.mp4)',
      'Páginas',
      'Resolução',
      'Data',
      'Tamanho (MB)',
      'Extensão do arquivo'
    ]);
    expect(boxes[5]).toBeDisabled();
    expect(boxes[0]).toBeEnabled();
    fireEvent.click(boxes[3]);
    expect(onOptionChange).toHaveBeenCalledWith('byDate', true);
  });
});

describe('FolderPathsPanel', () => {
  it('abrevia o destino de 73 caracteres e mantém o caminho completo no title', () => {
    const long = `C:\\${'x'.repeat(70)}`;
    render(
      <FolderPathsPanel
        sourceLabel="Origem"
        destinationLabel="Destino"
        sourcePath="C:\\origem"
        destinationPath={long}
        destinationSelectHintPrefix=""
        destinationSelectHintAction="Selecione"
        onSelectDestinationFolder={vi.fn()}
      />
    );
    const shown = screen.getByTitle(long);
    expect(shown.textContent).toContain('...');
    expect(shown.textContent).toHaveLength(67);
  });
});
