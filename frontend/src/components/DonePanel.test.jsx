import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import DonePanel, { barWidth } from './DonePanel';

const pt = organizerCopy['pt-BR'];

const result = (overrides = {}) => ({
  movedFiles: 5,
  failedFiles: 0,
  destinationFolderPath: 'C:\\Users\\caio\\Organizados',
  folders: [
    { name: 'pdf', count: 4 },
    { name: '', count: 1 }
  ],
  otherFiles: 0,
  ...overrides
});

function renderDone(props = {}) {
  const handlers = { onOpenDestination: vi.fn(), onUndo: vi.fn(), onStartOver: vi.fn() };
  const view = render(
    <DonePanel labels={pt} result={result()} hasUndo isLoading={false} {...handlers} {...props} />
  );
  return { ...view, ...handlers };
}

const rows = (container) =>
  [...container.querySelectorAll('.st-done__row')].map((row) => [
    row.querySelector('.st-done__folder').textContent,
    row.querySelector('.st-done__count').textContent,
    row.querySelector('.st-done__bar').style.width
  ]);

describe('barWidth', () => {
  it.each([
    [0, 0, '0%'],
    [1, 4, '25%'],
    [4, 4, '100%']
  ])('%i de %i → %s', (count, max, expected) => {
    expect(barWidth(count, max)).toBe(expected);
  });
});

describe('DonePanel', () => {
  it('número em destaque, título e a pasta de destino', () => {
    renderDone();
    expect(screen.getByText('5')).toHaveClass('st-done__number');
    expect(screen.getByText(pt.doneTitle)).toBeInTheDocument();
    expect(screen.getByText(`${pt.doneIn} Organizados`)).toBeInTheDocument();
  });

  it.each([
    [1, pt.doneTitleOne],
    [2, pt.doneTitle]
  ])('%i arquivo(s): "%s"', (movedFiles, title) => {
    renderDone({ result: result({ movedFiles }) });
    expect(screen.getByText(title)).toBeInTheDocument();
  });

  it.each([
    [0, null],
    [1, `1 ${pt.doneFailedOne}`],
    [2, `2 ${pt.doneFailed}`]
  ])('%i falha(s): "%s"', (failedFiles, text) => {
    const { container } = renderDone({ result: result({ failedFiles }) });
    const err = container.querySelector('.st-done__sub--err');
    if (text) {
      expect(err).toHaveTextContent(text);
    } else {
      expect(err).toBeNull();
    }
  });

  it('uma barra por pasta, relativa à maior, com a raiz traduzida', () => {
    const { container } = renderDone();
    expect(rows(container)).toEqual([
      ['pdf', '4', '100%'],
      [pt.previewRoot, '1', '25%']
    ]);
  });

  it('com o critério Tipo, as barras mostram as categorias traduzidas', () => {
    const { container } = renderDone({
      result: result({ categoryFolders: true, folders: [{ name: 'documents', count: 3 }] })
    });
    expect(rows(container)[0][0]).toBe('Documentos');
    expect(container.querySelector('.st-done__folder')).toHaveAttribute('title', 'documents');
  });

  it('"Outras" vira mais uma barra', () => {
    const { container } = renderDone({ result: result({ otherFiles: 8 }) });
    expect(rows(container).at(-1)).toEqual([pt.previewOthers, '8', '100%']);
  });

  it('ao aparecer, o foco vai para o resumo', () => {
    const { container } = renderDone();
    expect(container.querySelector('.st-done__head')).toHaveFocus();
  });

  it('os três botões chamam as ações', () => {
    const { onOpenDestination, onUndo, onStartOver } = renderDone();
    fireEvent.click(screen.getByRole('button', { name: pt.openDestination }));
    fireEvent.click(screen.getByRole('button', { name: pt.undo }));
    fireEvent.click(screen.getByRole('button', { name: pt.startOver }));
    expect(onOpenDestination).toHaveBeenCalled();
    expect(onUndo).toHaveBeenCalled();
    expect(onStartOver).toHaveBeenCalled();
  });

  it.each([
    [true, false, false],
    [false, false, true],
    [true, true, true]
  ])(
    'desfazer disponível = %s, ocupado = %s: Desfazer desabilitado = %s',
    (hasUndo, isLoading, disabled) => {
      renderDone({ hasUndo, isLoading });
      expect(screen.getByRole('button', { name: pt.undo }).disabled).toBe(disabled);
    }
  );
});
