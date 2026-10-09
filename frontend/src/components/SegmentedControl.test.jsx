import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import SegmentedControl from './SegmentedControl';

const options = [
  { value: 'dark', label: 'Escuro' },
  { value: 'light', label: 'Claro' }
];

describe('SegmentedControl', () => {
  it.each([
    ['dark', 'true', 'false'],
    ['light', 'false', 'true']
  ])('valor %s: aria-pressed Escuro %s, Claro %s', (value, dark, light) => {
    render(<SegmentedControl label="Tema" options={options} value={value} onChange={vi.fn()} />);
    expect(screen.getByRole('group', { name: 'Tema' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Escuro' })).toHaveAttribute('aria-pressed', dark);
    expect(screen.getByRole('button', { name: 'Claro' })).toHaveAttribute('aria-pressed', light);
  });

  it('clicar pede o valor da opção', () => {
    const onChange = vi.fn();
    render(<SegmentedControl label="Tema" options={options} value="dark" onChange={onChange} />);
    fireEvent.click(screen.getByRole('button', { name: 'Claro' }));
    expect(onChange).toHaveBeenCalledWith('light');
  });
});
