import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import OrganizerSettingsPanel from './OrganizerSettingsPanel';

const labels = organizerCopy['pt-BR'];

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

  it.each([
    [true, 'opacity-100'],
    [false, 'pointer-events-none']
  ])('aberto = %s: painel com %s; a engrenagem alterna', (isOpen, cls) => {
    const onToggle = vi.fn();
    const { container } = render(
      <OrganizerSettingsPanel
        isOpen={isOpen}
        labels={labels}
        criteria={[]}
        onToggle={onToggle}
        onCriterionChange={vi.fn()}
      />
    );
    expect(container.querySelector('h3').parentElement.className).toContain(cls);
    fireEvent.click(screen.getByRole('button', { name: labels.settingsTitle }));
    expect(onToggle).toHaveBeenCalled();
  });
});
