import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import PageHead from './PageHead';

describe('PageHead', () => {
  it('título como h1, subtítulo e as ações à direita', () => {
    render(
      <PageHead title="Organizar pasta" subtitle="Selecione origem e destino.">
        <button type="button">Notificações</button>
      </PageHead>
    );
    expect(screen.getByRole('heading', { level: 1, name: 'Organizar pasta' })).toBeInTheDocument();
    expect(screen.getByText('Selecione origem e destino.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Notificações' })).toBeInTheDocument();
  });
});
