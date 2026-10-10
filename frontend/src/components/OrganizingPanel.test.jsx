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
      progress={{ done: 0, total: 0, file: '', folder: '', categoryFolders: false, ...progress }}
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
    ['', '', false, null],
    ['a.pdf', 'pdf', false, 'pdf'],
    ['b.txt', '', false, pt.previewRoot],
    // Regressão (#84): a categoria aparecia como a chave interna (images).
    ['c.jpg', 'images', true, 'Imagens']
  ])('arquivo "%s" → pasta "%s" (categorias = %s)', (file, folder, categoryFolders, shown) => {
    const { container } = renderPanel({ done: 1, total: 2, file, folder, categoryFolders });
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

  it('ao aparecer, o foco vai para o Cancelar', () => {
    renderPanel({ total: 0 });
    expect(screen.getByRole('button', { name: pt.cancel })).toHaveFocus();
  });

  it('desfazendo: mesma tela no sentido inverso, do destino para a origem', () => {
    const { container } = renderPanel(
      { done: 1, total: 3, file: 'a.pdf', folder: '' },
      { mode: 'restore', destinationFolderPath: 'D:\\Organizados' }
    );
    expect(screen.getByText(pt.restoring)).toBeInTheDocument();
    expect(screen.getByRole('progressbar', { name: pt.restoreProgressLabel })).toBeInTheDocument();
    expect(container.querySelector('.st-flow__labels')).toHaveTextContent(
      'Organizados · 2Downloads · 1'
    );
    expect(container.querySelector('.st-running__now')).toHaveTextContent(
      `${pt.restoreMoving} a.pdf → ${pt.previewRoot}`
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
