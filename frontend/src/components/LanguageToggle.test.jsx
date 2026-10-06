import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import LanguageToggle from './LanguageToggle';

describe('LanguageToggle', () => {
  it.each([
    ['pt-BR', 'true', 'false'],
    ['en', 'false', 'true']
  ])('com %s, informa qual idioma está ativo', (language, ptPressed, enPressed) => {
    render(<LanguageToggle language={language} onChange={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'PT-BR' })).toHaveAttribute(
      'aria-pressed',
      ptPressed
    );
    expect(screen.getByRole('button', { name: 'EN' })).toHaveAttribute('aria-pressed', enPressed);
  });

  // Regressão (#60): a bandeira repetia o rótulo ("PT-BR PT-BR" no leitor de tela).
  it('a bandeira é decorativa e o nome do botão é só o rótulo', () => {
    const { container } = render(<LanguageToggle language="pt-BR" onChange={vi.fn()} />);
    container.querySelectorAll('img').forEach((img) => expect(img).toHaveAttribute('alt', ''));
    expect(screen.queryByRole('button', { name: 'PT-BR PT-BR' })).not.toBeInTheDocument();
  });

  it('clicar troca o idioma', () => {
    const onChange = vi.fn();
    render(<LanguageToggle language="pt-BR" onChange={onChange} />);
    fireEvent.click(screen.getByRole('button', { name: 'EN' }));
    expect(onChange).toHaveBeenCalledWith('en');
  });
});
