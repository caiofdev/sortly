import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import ProgressBlocks, { BLOCKS, filledBlocks } from './ProgressBlocks';

describe('filledBlocks', () => {
  it.each([
    [0, 0, 0],
    [0, 128, 0],
    [3, 128, 0],
    [4, 128, 1],
    [127, 128, 31],
    [128, 128, BLOCKS]
  ])('%i de %i → %i blocos', (done, total, expected) => {
    expect(filledBlocks(done, total)).toBe(expected);
  });
});

describe('ProgressBlocks', () => {
  it('barra acessível com 32 blocos e os cheios marcados', () => {
    const { container } = render(<ProgressBlocks label="Progresso" done={64} total={128} />);
    const bar = screen.getByRole('progressbar', { name: 'Progresso' });
    expect(bar).toHaveAttribute('aria-valuenow', '64');
    expect(bar).toHaveAttribute('aria-valuemax', '128');
    expect(container.querySelectorAll('.st-blocks__b')).toHaveLength(BLOCKS);
    expect(container.querySelectorAll('.st-blocks__b--on')).toHaveLength(16);
  });

  it('total 1 concluído: todos os blocos cheios', () => {
    const { container } = render(<ProgressBlocks label="Progresso" done={1} total={1} />);
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuemax', '1');
    expect(container.querySelectorAll('.st-blocks__b--on')).toHaveLength(BLOCKS);
  });

  it('sem total: barra indeterminada, sem valor', () => {
    render(<ProgressBlocks label="Progresso" done={0} total={0} />);
    const bar = screen.getByRole('progressbar', { name: 'Progresso' });
    expect(bar).not.toHaveAttribute('aria-valuenow');
    expect(bar).not.toHaveAttribute('aria-valuemax');
  });
});
