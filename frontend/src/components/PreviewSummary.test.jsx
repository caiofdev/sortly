import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import PreviewSummary from './PreviewSummary';

const pt = organizerCopy['pt-BR'];
const preview = (overrides = {}) => ({
  status: 'ready',
  totalFiles: 3,
  folders: [],
  otherFiles: 0,
  ...overrides
});

function renderPreview(value) {
  return render(<PreviewSummary labels={pt} preview={value} />);
}

const chips = (container) =>
  [...container.querySelectorAll('.st-chip')].map((chip) => chip.textContent);

describe('PreviewSummary', () => {
  it('sem origem: região anunciável vazia', () => {
    const { container } = renderPreview(preview({ status: '' }));
    expect(container.firstChild).toBeEmptyDOMElement();
    expect(container.firstChild).toHaveAttribute('aria-live', 'polite');
  });

  it('contando: mostra o aviso, sem chips', () => {
    const { container } = renderPreview(preview({ status: 'loading' }));
    expect(screen.getByText(pt.previewCounting)).toBeInTheDocument();
    expect(chips(container)).toEqual([]);
  });

  it.each([
    [0, pt.previewFound],
    [1, pt.previewFoundOne],
    [2, pt.previewFound]
  ])('%i arquivo(s): "%s"', (totalFiles, text) => {
    const { container } = renderPreview(preview({ totalFiles }));
    expect(container.firstChild).toHaveTextContent(`${totalFiles} ${text}`);
  });

  it('com o critério Tipo, as categorias aparecem traduzidas', () => {
    const { container } = renderPreview(
      preview({ categoryFolders: true, folders: [{ name: 'images', count: 2 }] })
    );
    expect(chips(container)).toEqual(['Imagens 2']);
  });

  it('um chip por pasta, com a raiz do destino traduzida', () => {
    const { container } = renderPreview(
      preview({
        folders: [
          { name: 'pdf', count: 2 },
          { name: '', count: 1 }
        ]
      })
    );
    expect(chips(container)).toEqual(['pdf 2', `${pt.previewRoot} 1`]);
  });

  it.each([
    [0, []],
    [1, [`${pt.previewOthers} 1`]]
  ])('%i em outras pastas: chips extras %j', (otherFiles, extra) => {
    const { container } = renderPreview(
      preview({ folders: [{ name: 'pdf', count: 2 }], otherFiles })
    );
    expect(chips(container)).toEqual(['pdf 2', ...extra]);
  });
});
