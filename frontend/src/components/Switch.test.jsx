import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import Switch from './Switch';

describe('Switch', () => {
  it.each([
    [false, 'false', true],
    [true, 'true', false]
  ])('ligado = %s: aria-checked %s; clicar pede %s', (checked, ariaChecked, next) => {
    const onChange = vi.fn();
    render(<Switch checked={checked} label="Data" onChange={onChange} />);
    const control = screen.getByRole('switch', { name: 'Data' });
    expect(control).toHaveAttribute('aria-checked', ariaChecked);
    fireEvent.click(control);
    expect(onChange).toHaveBeenCalledWith(next);
  });

  it('desabilitado não muda', () => {
    const onChange = vi.fn();
    render(<Switch checked label="Extensão" disabled onChange={onChange} />);
    fireEvent.click(screen.getByRole('switch', { name: 'Extensão' }));
    expect(onChange).not.toHaveBeenCalled();
  });
});
