// Teste de fumaça: a tela principal monta e consulta o estado da última organização.

import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import App from './App';

describe('App', () => {
  beforeEach(() => {
    window.go = {
      app: {
        App: {
          GetLastOrganizationState: vi.fn().mockResolvedValue({
            hasUndo: false,
            sourceFolderPath: '',
            destinationFolderPath: ''
          })
        }
      }
    };
    window.runtime = { OnFileDrop: vi.fn(), OnFileDropOff: vi.fn() };
  });

  it('renderiza a tela principal em português por padrão', async () => {
    render(<App />);

    expect(screen.getByRole('heading', { name: 'Sortly' })).toBeInTheDocument();
    expect(screen.getByText('Arraste e solte uma pasta ou arquivo aqui')).toBeInTheDocument();
    await vi.waitFor(() => expect(window.go.app.App.GetLastOrganizationState).toHaveBeenCalled());
  });

  it('usa o idioma salvo', () => {
    window.localStorage.setItem('sortly.language', 'en');

    render(<App />);

    expect(screen.getByText('Drag and drop a folder or file here')).toBeInTheDocument();
  });
});
