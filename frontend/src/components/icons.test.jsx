import { describe, expect, it } from 'vitest';
import { render } from '@testing-library/react';
import { BellIcon, CheckIcon, DropFolderIcon } from './icons';

describe('icons', () => {
  it.each([
    ['padrão', BellIcon, '2', null],
    ['da dropzone', DropFolderIcon, '1.75', 'st-drop__icon'],
    ['do toast', CheckIcon, '2.5', null]
  ])('ícone %s: traço %s, decorativo', (_name, Component, stroke, className) => {
    const { container } = render(<Component />);
    const svg = container.querySelector('svg');
    expect(svg).toHaveAttribute('stroke-width', stroke);
    expect(svg).toHaveAttribute('aria-hidden', 'true');
    expect(svg.getAttribute('class')).toBe(className);
  });
});
