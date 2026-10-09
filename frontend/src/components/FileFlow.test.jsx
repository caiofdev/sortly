import { describe, expect, it } from 'vitest';
import { render } from '@testing-library/react';
import FileFlow from './FileFlow';

function renderFlow(done, total) {
  return render(
    <FileFlow done={done} total={total} sourceName="Downloads" destinationName="Organizados" />
  );
}

const peeks = (container) =>
  [...container.querySelectorAll('.st-flow__peek')].map((el) => Number.parseInt(el.style.left, 10));

describe('FileFlow', () => {
  it('é decorativa: escondida do leitor de tela, com seis arquivos em voo', () => {
    const { container } = renderFlow(0, 10);
    expect(container.querySelector('.st-flow')).toHaveAttribute('aria-hidden', 'true');
    expect(container.querySelectorAll('.st-flow__file')).toHaveLength(6);
  });

  it.each([
    ['sem total (preparando)', 0, 0, [22, 54]],
    ['no começo', 0, 8, [22, 54]],
    ['passou de 1/8', 2, 8, [22, 54, 440]],
    ['na metade', 4, 8, [54, 440]],
    ['passou de 5/8', 6, 8, [54, 440, 474]],
    ['no fim', 8, 8, [440, 474]]
  ])('%s: arquivos espiando %j', (_name, done, total, expected) => {
    expect(peeks(renderFlow(done, total).container)).toEqual(expected);
  });

  it('as contagens de origem e destino seguem o progresso', () => {
    const { container } = renderFlow(3, 10);
    expect(container.querySelector('.st-flow__labels')).toHaveTextContent(
      'Downloads · 7Organizados · 3'
    );
  });
});
