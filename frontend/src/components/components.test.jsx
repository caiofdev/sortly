import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
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
  it('mostra os critérios na ordem recebida e desabilita o travado', () => {
    const onCriterionChange = vi.fn();
    const criteria = [
      'byDuration',
      'byPages',
      'byResolution',
      'byDate',
      'bySize',
      'byExtension'
    ].map((key) => ({ key, enabled: key === 'byExtension', locked: key === 'byExtension' }));
    render(
      <OrganizerSettingsPanel
        isOpen
        labels={labels}
        criteria={criteria}
        onToggle={vi.fn()}
        onCriterionChange={onCriterionChange}
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
    expect(boxes[5]).toBeChecked();
    expect(boxes[3]).not.toBeChecked();
    expect(boxes[0]).toBeEnabled();
    fireEvent.click(boxes[3]);
    expect(onCriterionChange).toHaveBeenCalledWith('byDate', true);
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
