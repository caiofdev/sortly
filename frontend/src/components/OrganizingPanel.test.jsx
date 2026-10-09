import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import OrganizingPanel, { folderName } from './OrganizingPanel';

const pt = organizerCopy['pt-BR'];

function renderPanel(progress, props = {}) {
  const onCancel = vi.fn();
  const view = render(
    <OrganizingPanel
      labels={pt}
      progress={{ done: 0, total: 0, file: '', folder: '', ...progress }}
      sourceFolderPath="C:\Users\caio\Downloads"
      destinationFolderPath=""
      onCancel={onCancel}
      {...props}
    />
  );
  return { ...view, onCancel };
}

describe('folderName', () => {
  it.each([
    ['C:\\Users\\caio\\Downloads', 'Downloads'],
    ['C:\\Users\\caio\\Downloads\\', 'Downloads'],
    ['/home/caio/Downloads', 'Downloads'],
    ['', '']
  ])('%s → %s', (path, expected) => {
    expect(folderName(path)).toBe(expected);
  });
});

describe('OrganizingPanel', () => {
  it.each([
    [{ total: 0 }, pt.progressPreparing],
    [{ done: 3, total: 10 }, '3 / 10']
  ])('progresso %j: contagem "%s"', (progress, text) => {
    renderPanel(progress);
    expect(screen.getByText(text)).toBeInTheDocument();
  });

  it.each([
    ['', '', null],
    ['a.pdf', 'pdf', 'pdf'],
    ['b.txt', '', pt.previewRoot]
  ])('arquivo "%s" → pasta "%s"', (file, folder, shown) => {
    const { container } = renderPanel({ done: 1, total: 2, file, folder });
    const now = container.querySelector('.st-running__now');
    if (shown) {
      expect(now).toHaveTextContent(`${pt.moving} ${file} → ${shown}`);
    } else {
      expect(now).toBeEmptyDOMElement();
    }
  });

  it('sem destino, a pasta de destino é a própria origem', () => {
    const { container } = renderPanel({ done: 1, total: 2 });
    expect(container.querySelector('.st-flow__labels')).toHaveTextContent(
      'Downloads · 1Downloads · 1'
    );
  });

  it('Cancelar pede o cancelamento', () => {
    const { container, onCancel } = renderPanel(
      { done: 1, total: 2 },
      { destinationFolderPath: 'D:\\Organizados' }
    );
    expect(container.querySelector('.st-flow__labels')).toHaveTextContent('Organizados · 1');
    fireEvent.click(screen.getByRole('button', { name: pt.cancel }));
    expect(onCancel).toHaveBeenCalled();
  });
});
