import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import LanguageToggle from './LanguageToggle';

describe('LanguageToggle', () => {
  it.each([
    ['pt-BR', 'true', 'false'],
    ['en', 'false', 'true']
  ])('com %s, informa qual idioma está ativo', (language, ptPressed, enPressed) => {
    render(<LanguageToggle language={language} label="Idioma" onChange={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'PT-BR' })).toHaveAttribute(
      'aria-pressed',
      ptPressed
    );
    expect(screen.getByRole('button', { name: 'EN' })).toHaveAttribute('aria-pressed', enPressed);
  });

  // Regressão (#60): o leitor de tela lia "PT-BR PT-BR" com a bandeira.
  it('o nome do botão é só o rótulo, dentro de um grupo com nome', () => {
    render(<LanguageToggle language="pt-BR" label="Idioma" onChange={vi.fn()} />);
    expect(screen.getByRole('group', { name: 'Idioma' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'PT-BR PT-BR' })).not.toBeInTheDocument();
  });

  it('clicar troca o idioma', () => {
    const onChange = vi.fn();
    render(<LanguageToggle language="pt-BR" label="Idioma" onChange={onChange} />);
    fireEvent.click(screen.getByRole('button', { name: 'EN' }));
    expect(onChange).toHaveBeenCalledWith('en');
  });
});
