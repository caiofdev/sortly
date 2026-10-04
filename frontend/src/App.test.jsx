import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import App from './App';

const criteria = [
  { key: 'byDuration', enabled: false, locked: false },
  { key: 'byPages', enabled: false, locked: false },
  { key: 'byResolution', enabled: false, locked: false },
  { key: 'byDate', enabled: false, locked: false },
  { key: 'bySize', enabled: false, locked: false },
  { key: 'byExtension', enabled: true, locked: true }
];

function mockBackend(language) {
  window.go = {
    app: {
      App: {
        GetLastOrganizationState: vi.fn().mockResolvedValue({
          hasUndo: false,
          sourceFolderPath: '',
          destinationFolderPath: ''
        }),
        GetSettings: vi.fn().mockResolvedValue({ language, criteria })
      }
    }
  };
}

describe('App', () => {
  beforeEach(() => {
    window.runtime = { OnFileDrop: vi.fn(), OnFileDropOff: vi.fn() };
  });

  it('renderiza a tela principal em português por padrão', async () => {
    mockBackend('pt-BR');
    render(<App />);

    expect(screen.getByRole('heading', { name: 'Sortly' })).toBeInTheDocument();
    expect(screen.getByText('Arraste e solte uma pasta ou arquivo aqui')).toBeInTheDocument();
    await vi.waitFor(() => expect(window.go.app.App.GetLastOrganizationState).toHaveBeenCalled());
  });

  it('usa o idioma salvo no backend', async () => {
    mockBackend('en');
    render(<App />);

    expect(await screen.findByText('Drag and drop a folder or file here')).toBeInTheDocument();
  });
});
